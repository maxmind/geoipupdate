package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/maxmind/geoipupdate/v7/internal/geoipupdate"
)

func TestUpdater(t *testing.T) {
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
	err = os.WriteFile(dbFile, []byte("old edition-2 content"), os.ModePerm)
	require.NoError(t, err)

	lastModified, err := time.ParseInLocation("2006-01-02", "2024-02-23", time.UTC)
	require.NoError(t, err)

	// mock metadata handler.
	metadataHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		queryParams, err := url.ParseQuery(r.URL.RawQuery)
		require.NoError(t, err)

		var jsonData string
		switch queryParams.Get("edition_id") {
		case "edition-1":
			jsonData = `{
    		"databases": [
    		    {
    		        "edition_id": "edition-1",
    		        "md5": "618dd27a10de24809ec160d6807f363f",
    		        "date": "2024-02-23"
    		    }
    		]
		}`
		case "edition-2":
			jsonData = `{
		    "databases": [
		        {
		            "edition_id": "edition-2",
		            "md5": "c9bbf7cb507370339633b44001bae038",
		            "date": "2024-02-23"
		        }
		    ]
		}`
		default:
			t.Error("unsupported edition in metadata request")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err = w.Write([]byte(jsonData))
		require.NoError(t, err)
	})

	// mock download handler.
	downloadHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.Split(r.URL.Path, "/")[3] // extract the edition-id.

		var buf bytes.Buffer
		gw := gzip.NewWriter(&buf)
		tw := tar.NewWriter(gw)

		content := name + " content"
		header := &tar.Header{
			Name: name + ".mmdb",
			Size: int64(len(content)),
		}

		err := tw.WriteHeader(header)
		require.NoError(t, err)
		_, err = tw.Write([]byte(content))
		require.NoError(t, err)

		require.NoError(t, tw.Close())
		require.NoError(t, gw.Close())

		w.Header().Set("Content-Type", "application/gzip")
		w.Header().Set("Content-Disposition", "attachment; filename=test.tar.gz")
		w.Header().Set("Last-Modified", lastModified.Format(time.RFC1123))
		_, err = io.Copy(w, &buf)
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

	config := &geoipupdate.Config{
		AccountID:         123,
		DatabaseDirectory: tempDir,
		EditionIDs:        []string{"edition-1", "edition-2"},
		LicenseKey:        "testing",
		LockFile:          filepath.Join(tempDir, ".geoipupdate.lock"),
		URL:               server.URL,
		Parallelism:       1,
		Output:            true,
	}

	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	updater, err := geoipupdate.NewUpdater(config)
	require.NoError(t, err)

	err = updater.Run(context.Background())
	require.NoError(t, err, "run successfully")

	w.Close()
	out, err := io.ReadAll(r)
	require.NoError(t, err)
	//nolint:lll
	expectedOutput := `\[{"edition_id":"edition\-1","old_hash":"618dd27a10de24809ec160d6807f363f","new_hash":"618dd27a10de24809ec160d6807f363f","checked_at":\d+},{"edition_id":"edition\-2","old_hash":"2242f06b3b2d147987b67017cb7a5ab8","new_hash":"c9bbf7cb507370339633b44001bae038","modified_at":1708646400,"checked_at":\d+}]`
	require.Regexp(t, expectedOutput, string(out))

	for _, editionID := range config.EditionIDs {
		path := filepath.Join(config.DatabaseDirectory, editionID+".mmdb")
		buf, err := os.ReadFile(filepath.Clean(path))
		require.NoError(t, err, "read file")
		require.Equal(
			t,
			editionID+" content",
			string(buf),
			"correct database",
		)
	}
}
