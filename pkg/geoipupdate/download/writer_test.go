package download

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestDownloadEdition checks the database download functionality.
func TestDownloadEdition(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "db")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	ctx := context.Background()
	now, err := ParseTime("2024-02-23")
	require.NoError(t, err)

	edition := Metadata{
		EditionID: "edition-1",
		Date:      "2024-02-02",
		MD5:       "618dd27a10de24809ec160d6807f363f",
	}

	dbContent := "edition-1 content"

	tests := []struct {
		description      string
		preserveFileTime bool
		server           func(t *testing.T) *httptest.Server
		checkResult      func(t *testing.T, err error)
	}{
		{
			description:      "successful download",
			preserveFileTime: false,
			server: func(t *testing.T) *httptest.Server {
				var buf bytes.Buffer
				gw := gzip.NewWriter(&buf)
				tw := tar.NewWriter(gw)

				header := &tar.Header{
					Name: "edition-1" + Extension,
					Size: int64(len(dbContent)),
				}

				err = tw.WriteHeader(header)
				require.NoError(t, err)
				_, err = tw.Write([]byte(dbContent))
				require.NoError(t, err)

				require.NoError(t, tw.Close())
				require.NoError(t, gw.Close())

				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.Header().Set("Content-Type", "application/gzip")
					w.Header().Set("Content-Disposition", "attachment; filename=test.tar.gz")
					_, err := io.Copy(w, &buf)
					require.NoError(t, err)
				}))

				return server
			},
			checkResult: func(t *testing.T, err error) {
				require.NoError(t, err)

				dbFile := filepath.Join(tempDir, edition.EditionID+Extension)

				//nolint:gosec // we need to read the content of the file in this test.
				fileContent, err := os.ReadFile(dbFile)
				require.NoError(t, err)
				require.Equal(t, dbContent, string(fileContent))

				database, err := os.Stat(dbFile)
				require.NoError(t, err)
				require.GreaterOrEqual(t, database.ModTime(), now)
			},
		},
		{
			description:      "successful download - preserve time",
			preserveFileTime: true,
			server: func(t *testing.T) *httptest.Server {
				var buf bytes.Buffer
				gw := gzip.NewWriter(&buf)
				tw := tar.NewWriter(gw)

				header := &tar.Header{
					Name: "edition-1" + Extension,
					Size: int64(len(dbContent)),
				}

				err = tw.WriteHeader(header)
				require.NoError(t, err)
				_, err = tw.Write([]byte(dbContent))
				require.NoError(t, err)

				require.NoError(t, tw.Close())
				require.NoError(t, gw.Close())

				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.Header().Set("Content-Type", "application/gzip")
					w.Header().Set("Content-Disposition", "attachment; filename=test.tar.gz")
					_, err := io.Copy(w, &buf)
					require.NoError(t, err)
				}))

				return server
			},
			checkResult: func(t *testing.T, err error) {
				require.NoError(t, err)

				dbFile := filepath.Join(tempDir, edition.EditionID+Extension)

				//nolint:gosec // we need to read the content of the file in this test.
				fileContent, err := os.ReadFile(dbFile)
				require.NoError(t, err)
				require.Equal(t, dbContent, string(fileContent))

				modTime, err := ParseTime(edition.Date)
				require.NoError(t, err)

				database, err := os.Stat(dbFile)
				require.NoError(t, err)
				require.Equal(t, modTime, database.ModTime())
			},
		},
		{
			description:      "server error",
			preserveFileTime: false,
			server: func(_ *testing.T) *httptest.Server {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				}))
				return server
			},
			checkResult: func(t *testing.T, err error) {
				require.Error(t, err)
				require.Regexp(t, "^requesting edition: received: HTTP status code '500'", err.Error())
			},
		},
		{
			description:      "wrong file format",
			preserveFileTime: false,
			server: func(t *testing.T) *httptest.Server {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					jsonData := `{"message": "Hello, world!", "status": "ok"}`
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					_, err := w.Write([]byte(jsonData))
					require.NoError(t, err)
				}))
				return server
			},
			checkResult: func(t *testing.T, err error) {
				require.Error(t, err)
				require.Regexp(t, "^encountered an error creating GZIP reader: gzip: invalid header", err.Error())
			},
		},
		{
			description:      "empty tar archive",
			preserveFileTime: false,
			server: func(t *testing.T) *httptest.Server {
				var buf bytes.Buffer
				gw := gzip.NewWriter(&buf)
				tw := tar.NewWriter(gw)
				require.NoError(t, tw.Close())
				require.NoError(t, gw.Close())

				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.Header().Set("Content-Type", "application/gzip")
					w.Header().Set("Content-Disposition", "attachment; filename=test.tar.gz")
					_, err := io.Copy(w, &buf)
					require.NoError(t, err)
				}))

				return server
			},
			checkResult: func(t *testing.T, err error) {
				require.Error(t, err)
				require.Regexp(t, "^tar archive does not contain an mmdb file", err.Error())
			},
		},
		{
			description:      "tar does not contain an mmdb file",
			preserveFileTime: false,
			server: func(t *testing.T) *httptest.Server {
				var buf bytes.Buffer
				gw := gzip.NewWriter(&buf)
				tw := tar.NewWriter(gw)

				header := &tar.Header{
					Name: "edition-1.zip",
					Size: int64(len(dbContent)),
				}

				err = tw.WriteHeader(header)
				require.NoError(t, err)
				_, err = tw.Write([]byte(dbContent))
				require.NoError(t, err)

				require.NoError(t, tw.Close())
				require.NoError(t, gw.Close())

				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.Header().Set("Content-Type", "application/gzip")
					w.Header().Set("Content-Disposition", "attachment; filename=test.tar.gz")
					_, err := io.Copy(w, &buf)
					require.NoError(t, err)
				}))

				return server
			},
			checkResult: func(t *testing.T, err error) {
				require.Error(t, err)
				require.Regexp(t, "^tar archive does not contain an mmdb file", err.Error())
			},
		},
		{
			description:      "mmdb hash does not match metadata",
			preserveFileTime: false,
			server: func(t *testing.T) *httptest.Server {
				var buf bytes.Buffer
				gw := gzip.NewWriter(&buf)
				tw := tar.NewWriter(gw)

				header := &tar.Header{
					Name: "edition-1" + Extension,
					Size: int64(len(dbContent) - 1),
				}

				err = tw.WriteHeader(header)
				require.NoError(t, err)
				_, err = tw.Write([]byte(dbContent[:len(dbContent)-1]))
				require.NoError(t, err)

				require.NoError(t, tw.Close())
				require.NoError(t, gw.Close())

				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.Header().Set("Content-Type", "application/gzip")
					w.Header().Set("Content-Disposition", "attachment; filename=test.tar.gz")
					_, err := io.Copy(w, &buf)
					require.NoError(t, err)
				}))

				return server
			},
			checkResult: func(t *testing.T, err error) {
				require.Error(t, err)
				//nolint:lll
				require.Regexp(t, "^validating hash for edition-1: md5 of new database .* does not match expected md5", err.Error())
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			server := test.server(t)
			defer server.Close()

			d := Download{
				client:            http.DefaultClient,
				databaseDir:       tempDir,
				now:               func() time.Time { return now },
				preserveFileTimes: test.preserveFileTime,
				url:               server.URL,
			}

			err = d.DownloadEdition(ctx, edition)
			test.checkResult(t, err)
		})
	}
}
