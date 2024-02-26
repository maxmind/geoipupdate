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

	"github.com/maxmind/geoipupdate/v6/pkg/geoipupdate/download"
	"github.com/stretchr/testify/assert"
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
	dbFile := filepath.Join(tempDir, edition+download.Extension)
	// equivalent MD5: 618dd27a10de24809ec160d6807f363f
	err = os.WriteFile(dbFile, []byte("edition-1 content"), os.ModePerm)
	require.NoError(t, err)

	edition = "edition-2"
	dbFile = filepath.Join(tempDir, edition+download.Extension)
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
			Name: name + download.Extension,
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
	conf := &Config{
		AccountID:         0,              // AccountID is not relevant for this test.
		LicenseKey:        "000000000001", // LicenseKey is not relevant for this test.
		DatabaseDirectory: tempDir,
		EditionIDs:        []string{"edition-1", "edition-2", "edition-3"},
		LockFile:          filepath.Clean(filepath.Join(tempDir, ".geoipupdate.lock")),
		URL:               server.URL,
		RetryFor:          0,
		Parallelism:       1,
		PreserveFileTimes: true,
	}

	logOutput := &bytes.Buffer{}
	log.SetOutput(logOutput)

	client, err := NewClient(conf)
	require.NoError(t, err)

	// download updates.
	err = client.Run(ctx)
	require.NoError(t, err)

	assert.Equal(t, "", logOutput.String(), "no logged output")

	// edition-1 file hasn't been modified.
	dbFile = filepath.Join(tempDir, "edition-1"+download.Extension)
	//nolint:gosec // we need to read the content of the file in this test.
	fileContent, err := os.ReadFile(dbFile)
	require.NoError(t, err)
	require.Equal(t, "edition-1 content", string(fileContent))
	database, err := os.Stat(dbFile)
	require.NoError(t, err)
	require.LessOrEqual(t, testDate, database.ModTime())

	// edition-2 file has been updated.
	dbFile = filepath.Join(tempDir, "edition-2"+download.Extension)
	//nolint:gosec // we need to read the content of the file in this test.
	fileContent, err = os.ReadFile(dbFile)
	require.NoError(t, err)
	require.Equal(t, "new edition-2 content", string(fileContent))
	modTime, err := download.ParseTime("2024-02-23")
	require.NoError(t, err)
	database, err = os.Stat(dbFile)
	require.NoError(t, err)
	require.Equal(t, modTime, database.ModTime())

	// edition-3 file has been downloaded.
	dbFile = filepath.Join(tempDir, "edition-3"+download.Extension)
	//nolint:gosec // we need to read the content of the file in this test.
	fileContent, err = os.ReadFile(dbFile)
	require.NoError(t, err)
	require.Equal(t, "new edition-3 content", string(fileContent))
	modTime, err = download.ParseTime("2024-02-02")
	require.NoError(t, err)
	database, err = os.Stat(dbFile)
	require.NoError(t, err)
	require.Equal(t, modTime, database.ModTime())
}
