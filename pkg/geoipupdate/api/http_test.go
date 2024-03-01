package api

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestGetMetadata checks the metadata fetching functionality.
func TestGetMetadata(t *testing.T) {
	tests := []struct {
		description      string
		preserveFileTime bool
		server           func(t *testing.T) *httptest.Server
		checkResult      func(t *testing.T, metadata []Metadata, err error)
	}{
		{
			description:      "successful request",
			preserveFileTime: false,
			server: func(t *testing.T) *httptest.Server {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					jsonData := `
{
    "databases": [
        { "edition_id": "edition-1", "md5": "123456", "date": "2024-02-23" },
        { "edition_id": "edition-2", "md5": "abc123", "date": "2024-02-23" },
        { "edition_id": "edition-3", "md5": "def456", "date": "2024-02-02" }
    ]
}
`
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					_, err := w.Write([]byte(jsonData))
					require.NoError(t, err)
				}))
				return server
			},
			checkResult: func(t *testing.T, metadata []Metadata, err error) {
				require.NoError(t, err)

				expectedMetadata := []Metadata{
					{EditionID: "edition-1", MD5: "123456", Date: "2024-02-23"},
					{EditionID: "edition-2", MD5: "abc123", Date: "2024-02-23"},
					{EditionID: "edition-3", MD5: "def456", Date: "2024-02-02"},
				}
				require.ElementsMatch(t, expectedMetadata, metadata)
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
			checkResult: func(t *testing.T, metadata []Metadata, err error) {
				require.Empty(t, metadata)
				require.Error(t, err)
				require.Regexp(t, "^requesting metadata: received: HTTP status code '500'", err.Error())
			},
		},
	}

	ctx := context.Background()

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			server := test.server(t)
			defer server.Close()

			h := NewHTTPDownloader(
				0,  // accountID is not relevant for this test.
				"", // licenseKey is not relevant for this test.
				http.DefaultClient,
				server.URL,
			)

			metadata, err := h.GetMetadata(ctx, []string{"edition-1", "edition-2", "edition-3"})
			test.checkResult(t, metadata, err)
		})
	}
}

// TestGetEdition checks the database download functionality.
func TestGetEdition(t *testing.T) {
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
		checkResult      func(t *testing.T, reader io.Reader, err error)
	}{
		{
			description:      "successful download",
			preserveFileTime: false,
			server: func(t *testing.T) *httptest.Server {
				var buf bytes.Buffer
				gw := gzip.NewWriter(&buf)
				tw := tar.NewWriter(gw)

				header := &tar.Header{
					Name: "edition-1.mmdb",
					Size: int64(len(dbContent)),
				}

				err := tw.WriteHeader(header)
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
			checkResult: func(t *testing.T, reader io.Reader, err error) {
				require.NoError(t, err)
				c, rerr := io.ReadAll(reader)
				require.NoError(t, rerr)
				require.Equal(t, dbContent, string(c))
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
			checkResult: func(t *testing.T, reader io.Reader, err error) {
				require.Nil(t, reader)
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
			checkResult: func(t *testing.T, reader io.Reader, err error) {
				require.Nil(t, reader)
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
			checkResult: func(t *testing.T, reader io.Reader, err error) {
				require.Nil(t, reader)
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

				err := tw.WriteHeader(header)
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
			checkResult: func(t *testing.T, reader io.Reader, err error) {
				require.Nil(t, reader)
				require.Error(t, err)
				require.Regexp(t, "^tar archive does not contain an mmdb file", err.Error())
			},
		},
	}

	ctx := context.Background()
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			server := test.server(t)
			defer server.Close()

			h := NewHTTPDownloader(
				0,  // accountID is not relevant for this test.
				"", // licenseKey is not relevant for this test.
				http.DefaultClient,
				server.URL,
			)

			reader, _, err := h.GetEdition(ctx, edition)
			test.checkResult(t, reader, err)
		})
	}
}
