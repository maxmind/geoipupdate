package database

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestRead checks the database download functionality.
func TestRead(t *testing.T) {
	edition := metadata{
		EditionID: "edition-1",
		Date:      "2024-02-02",
		MD5:       "123456",
	}
	dbContent := "edition-1 content"

	metadataHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		jsonData := `
{
    "databases": [
        { "edition_id": "edition-1", "md5": "618dd27a10de24809ec160d6807f363f", "date": "2024-02-23" }
    ]
}
`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(jsonData))
		require.NoError(t, err)
	})

	tests := []struct {
		description      string
		preserveFileTime bool
		server           func(t *testing.T) *httptest.Server
		checkResult      func(t *testing.T, resp *ReadResult, err error)
	}{
		{
			description:      "successful download",
			preserveFileTime: false,
			server: func(t *testing.T) *httptest.Server {
				downloadHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
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

					w.Header().Set("Content-Type", "application/gzip")
					w.Header().Set("Content-Disposition", "attachment; filename=test.tar.gz")
					_, err = io.Copy(w, &buf)
					require.NoError(t, err)
				})

				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if strings.HasPrefix(r.URL.Path, "/geoip/updates/metadata") {
						metadataHandler.ServeHTTP(w, r)
						return
					}

					downloadHandler.ServeHTTP(w, r)
				}))

				return server
			},
			checkResult: func(t *testing.T, resp *ReadResult, err error) {
				require.NoError(t, err)
				c, rerr := io.ReadAll(resp.reader)
				require.NoError(t, rerr)
				require.Equal(t, dbContent, string(c))
				require.Equal(t, edition.EditionID, resp.EditionID)
				require.Equal(t, edition.MD5, resp.OldHash)
				require.Equal(t, "618dd27a10de24809ec160d6807f363f", resp.NewHash)

				modifiedAt, err := parseTime("2024-02-23")
				require.NoError(t, err)
				require.Equal(t, modifiedAt, resp.ModifiedAt)
			},
		},
		{
			description:      "server error",
			preserveFileTime: false,
			server: func(_ *testing.T) *httptest.Server {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if strings.HasPrefix(r.URL.Path, "/geoip/updates/metadata") {
						metadataHandler.ServeHTTP(w, r)
						return
					}

					w.WriteHeader(http.StatusInternalServerError)
				}))
				return server
			},
			checkResult: func(t *testing.T, resp *ReadResult, err error) {
				require.Nil(t, resp)
				require.Error(t, err)
				require.Regexp(t, "^unexpected HTTP status code", err.Error())
			},
		},
		{
			description:      "wrong file format",
			preserveFileTime: false,
			server: func(t *testing.T) *httptest.Server {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if strings.HasPrefix(r.URL.Path, "/geoip/updates/metadata") {
						metadataHandler.ServeHTTP(w, r)
						return
					}

					jsonData := `{"message": "Hello, world!", "status": "ok"}`
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					_, err := w.Write([]byte(jsonData))
					require.NoError(t, err)
				}))
				return server
			},
			checkResult: func(t *testing.T, resp *ReadResult, err error) {
				require.Nil(t, resp)
				require.Error(t, err)
				require.Regexp(t, "^encountered an error creating GZIP reader", err.Error())
			},
		},
		{
			description:      "empty tar archive",
			preserveFileTime: false,
			server: func(t *testing.T) *httptest.Server {
				downloadHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					var buf bytes.Buffer
					gw := gzip.NewWriter(&buf)
					tw := tar.NewWriter(gw)
					require.NoError(t, tw.Close())
					require.NoError(t, gw.Close())

					w.Header().Set("Content-Type", "application/gzip")
					w.Header().Set("Content-Disposition", "attachment; filename=test.tar.gz")
					_, err := io.Copy(w, &buf)
					require.NoError(t, err)
				})

				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if strings.HasPrefix(r.URL.Path, "/geoip/updates/metadata") {
						metadataHandler.ServeHTTP(w, r)
						return
					}

					downloadHandler.ServeHTTP(w, r)
				}))

				return server
			},
			checkResult: func(t *testing.T, resp *ReadResult, err error) {
				require.Nil(t, resp)
				require.Error(t, err)
				require.Regexp(t, "^tar archive does not contain an mmdb file", err.Error())
			},
		},
		{
			description:      "tar does not contain an mmdb file",
			preserveFileTime: false,
			server: func(t *testing.T) *httptest.Server {
				downloadHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
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

					w.Header().Set("Content-Type", "application/gzip")
					w.Header().Set("Content-Disposition", "attachment; filename=test.tar.gz")
					_, err = io.Copy(w, &buf)
					require.NoError(t, err)
				})

				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if strings.HasPrefix(r.URL.Path, "/geoip/updates/metadata") {
						metadataHandler.ServeHTTP(w, r)
						return
					}

					downloadHandler.ServeHTTP(w, r)
				}))

				return server
			},
			checkResult: func(t *testing.T, resp *ReadResult, err error) {
				require.Nil(t, resp)
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

			r := NewHTTPReader(
				nil,        // request proxy.
				server.URL, // fixed, as the server is mocked above.
				10,         // fixed, as it's not valuable for the purpose of the test.
				"license",  // fixed, as it's not valuable for the purpose of the test.
				false,      // verbose
			)

			reader, err := r.get(ctx, edition.EditionID, edition.MD5)
			test.checkResult(t, reader, err)
		})
	}
}

// TestGetMetadata checks the metadata fetching functionality.
func TestGetMetadata(t *testing.T) {
	tests := []struct {
		description      string
		preserveFileTime bool
		server           func(t *testing.T) *httptest.Server
		checkResult      func(t *testing.T, receivedMetadata *metadata, err error)
	}{
		{
			description:      "successful request",
			preserveFileTime: false,
			server: func(t *testing.T) *httptest.Server {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					jsonData := `
{
    "databases": [
        { "edition_id": "edition-1", "md5": "123456", "date": "2024-02-23" }
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
			checkResult: func(t *testing.T, receivedMetadata *metadata, err error) {
				require.NoError(t, err)

				expectedMetadata := &metadata{
					EditionID: "edition-1", MD5: "123456", Date: "2024-02-23",
				}
				require.Equal(t, expectedMetadata, receivedMetadata)
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
			checkResult: func(t *testing.T, receivedMetadata *metadata, err error) {
				require.Nil(t, receivedMetadata)
				require.Error(t, err)
				require.Regexp(t, "^unexpected HTTP status code", err.Error())
			},
		},
	}

	ctx := context.Background()

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			server := test.server(t)
			defer server.Close()

			r := NewHTTPReader(
				nil,        // request proxy.
				server.URL, // fixed, as the server is mocked above.
				10,         // fixed, as it's not valuable for the purpose of the test.
				"license",  // fixed, as it's not valuable for the purpose of the test.
				false,      // verbose
			)

			result, err := r.getMetadata(ctx, "edition-1")
			test.checkResult(t, result, err)
		})
	}
}
