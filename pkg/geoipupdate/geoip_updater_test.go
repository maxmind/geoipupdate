package geoipupdate

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/maxmind/geoipupdate/v6/pkg/geoipupdate/api"
	"github.com/maxmind/geoipupdate/v6/pkg/geoipupdate/config"
	"github.com/maxmind/geoipupdate/v6/pkg/geoipupdate/lock"
	geoipupdatewriter "github.com/maxmind/geoipupdate/v6/pkg/geoipupdate/writer"
	"github.com/stretchr/testify/require"
)

// TestFullDownload runs an end to end test simulation.
func TestFullDownload(t *testing.T) {
	testDate := time.Now().Truncate(24 * time.Hour)

	// mock existing databases.
	tempDir, err := os.MkdirTemp("", "db")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	edition := "edition-1"
	dbFile := filepath.Join(tempDir, edition+".mmdb")
	// equivalent MD5: 618dd27a10de24809ec160d6807f363f
	err = os.WriteFile(dbFile, []byte("edition-1 content"), os.ModePerm)
	require.NoError(t, err)

	edition = "edition-2"
	dbFile = filepath.Join(tempDir, edition+".mmdb")
	err = os.WriteFile(dbFile, []byte("edition-2 content"), os.ModePerm)
	require.NoError(t, err)

	// mock metadata handler.
	metadataHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		jsonData := `
{
    "databases": [
        {
            "edition_id": "edition-1",
            "md5": "618dd27a10de24809ec160d6807f363f",
            "date": "2024-02-23"
        },
        {
            "edition_id": "edition-2",
            "md5": "9960e83daa34d69e9b58b375616e145b",
            "date": "2024-02-23"
        },
        {
            "edition_id": "edition-3",
            "md5": "08628247c1e8c1aa6d05ffc578fa09a8",
            "date": "2024-02-02"
        }
    ]
}
`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(jsonData))
		require.NoError(t, err)
	})

	// mock download handler.
	downloadHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.Split(r.URL.Path, "/")[3] // extract the edition-id.

		var buf bytes.Buffer
		gw := gzip.NewWriter(&buf)
		tw := tar.NewWriter(gw)

		content := "new " + name + " content"
		header := &tar.Header{
			Name: name + ".mmdb",
			Size: int64(len(content)),
		}

		err = tw.WriteHeader(header)
		require.NoError(t, err)
		_, err = tw.Write([]byte(content))
		require.NoError(t, err)

		require.NoError(t, tw.Close())
		require.NoError(t, gw.Close())

		w.Header().Set("Content-Type", "application/gzip")
		w.Header().Set("Content-Disposition", "attachment; filename=test.tar.gz")
		_, err := io.Copy(w, &buf)
		require.NoError(t, err)
	})

	// create test server.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/geoip/updates/metadata") {
			metadataHandler.ServeHTTP(w, r)
			return
		}

		if strings.HasPrefix(r.URL.Path, "/geoip/databases") {
			downloadHandler.ServeHTTP(w, r)
			return
		}

		http.NotFound(w, r)
	}))
	defer server.Close()

	ctx := context.Background()
	conf := &config.Config{
		AccountID:         0,              // AccountID is not relevant for this test.
		LicenseKey:        "000000000001", // LicenseKey is not relevant for this test.
		DatabaseDirectory: tempDir,
		EditionIDs:        []string{"edition-1", "edition-2", "edition-3"},
		LockFile:          filepath.Clean(filepath.Join(tempDir, ".geoipupdate.lock")),
		Output:            true,
		Parallelism:       1,
		PreserveFileTimes: true,
		RetryFor:          0,
		URL:               server.URL,
	}

	logOutput := &bytes.Buffer{}

	downloader := api.NewHTTPDownloader(
		conf.AccountID,
		conf.LicenseKey,
		http.DefaultClient,
		conf.URL,
	)

	writer := geoipupdatewriter.NewDiskWriter(conf.DatabaseDirectory, conf.PreserveFileTimes)

	locker, err := lock.NewFileLock(conf.LockFile)
	require.NoError(t, err)

	client := NewClient(conf, downloader, locker, writer)
	client.output = log.New(logOutput, "", 0)

	// download updates.
	err = client.Run(ctx)
	require.NoError(t, err)

	//nolint:lll
	expectedOutput := `[{"edition_id":"edition\-1","old_hash":"618dd27a10de24809ec160d6807f363f","new_hash":"618dd27a10de24809ec160d6807f363f","modified_at":1708646400,"checked_at":\d+},{"edition_id":"edition\-2","old_hash":"c9bbf7cb507370339633b44001bae038","new_hash":"9960e83daa34d69e9b58b375616e145b","modified_at":1708646400,"checked_at":\d+},{"edition_id":"edition\-3","old_hash":"00000000000000000000000000000000","new_hash":"08628247c1e8c1aa6d05ffc578fa09a8","modified_at":1706832000,"checked_at":\d+}]`
	require.Regexp(t, expectedOutput, logOutput.String())

	// edition-1 file hasn't been modified.
	dbFile = filepath.Join(tempDir, "edition-1.mmdb")
	//nolint:gosec // we need to read the content of the file in this test.
	fileContent, err := os.ReadFile(dbFile)
	require.NoError(t, err)
	require.Equal(t, "edition-1 content", string(fileContent))
	database, err := os.Stat(dbFile)
	require.NoError(t, err)
	require.LessOrEqual(t, testDate, database.ModTime())

	// edition-2 file has been updated.
	dbFile = filepath.Join(tempDir, "edition-2.mmdb")
	//nolint:gosec // we need to read the content of the file in this test.
	fileContent, err = os.ReadFile(dbFile)
	require.NoError(t, err)
	require.Equal(t, "new edition-2 content", string(fileContent))
	modTime, err := api.ParseTime("2024-02-23")
	require.NoError(t, err)
	database, err = os.Stat(dbFile)
	require.NoError(t, err)
	require.Equal(t, modTime, database.ModTime())

	// edition-3 file has been downloaded.
	dbFile = filepath.Join(tempDir, "edition-3.mmdb")
	//nolint:gosec // we need to read the content of the file in this test.
	fileContent, err = os.ReadFile(dbFile)
	require.NoError(t, err)
	require.Equal(t, "new edition-3 content", string(fileContent))
	modTime, err = api.ParseTime("2024-02-02")
	require.NoError(t, err)
	database, err = os.Stat(dbFile)
	require.NoError(t, err)
	require.Equal(t, modTime, database.ModTime())
}
