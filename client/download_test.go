package client

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
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDownload(t *testing.T) {
	edition := metadata{
		EditionID: "edition-1",
		Date:      "2024-02-02",
		MD5:       "123456",
	}
	dbContent := "edition-1 content"

	lastModified, err := time.ParseInLocation("2006-01-02", "2024-02-23", time.UTC)
	require.NoError(t, err)

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
		assert.NoError(t, err)
	})

	tests := []struct {
		description      string
		preserveFileTime bool
		server           func(t *testing.T) *httptest.Server
		checkResult      func(t *testing.T, res DownloadResponse, err error)
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
					if !assert.NoError(t, err) {
						return
					}
					_, err = tw.Write([]byte(dbContent))
					if !assert.NoError(t, err) {
						return
					}

					if !assert.NoError(t, tw.Close()) {
						return
					}
					if !assert.NoError(t, gw.Close()) {
						return
					}

					w.Header().Set("Content-Type", "application/gzip")
					w.Header().Set("Content-Disposition", "attachment; filename=test.tar.gz")
					w.Header().Set("Last-Modified", lastModified.Format(time.RFC1123))
					_, err = io.Copy(w, &buf)
					assert.NoError(t, err)
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
			checkResult: func(t *testing.T, res DownloadResponse, err error) {
				require.NoError(t, err)
				c, rerr := io.ReadAll(res.Reader)
				require.NoError(t, rerr)
				require.Equal(t, dbContent, string(c))
				require.Equal(t, "618dd27a10de24809ec160d6807f363f", res.MD5)
				require.Equal(t, lastModified, res.LastModified)
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
			checkResult: func(t *testing.T, _ DownloadResponse, err error) {
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
					assert.NoError(t, err)
				}))
				return server
			},
			checkResult: func(t *testing.T, _ DownloadResponse, err error) {
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
					if !assert.NoError(t, tw.Close()) {
						return
					}
					if !assert.NoError(t, gw.Close()) {
						return
					}

					w.Header().Set("Content-Type", "application/gzip")
					w.Header().Set("Content-Disposition", "attachment; filename=test.tar.gz")
					_, err := io.Copy(w, &buf)
					assert.NoError(t, err)
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
			checkResult: func(t *testing.T, _ DownloadResponse, err error) {
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
					if !assert.NoError(t, err) {
						return
					}
					_, err = tw.Write([]byte(dbContent))
					if !assert.NoError(t, err) {
						return
					}

					if !assert.NoError(t, tw.Close()) {
						return
					}
					if !assert.NoError(t, gw.Close()) {
						return
					}

					w.Header().Set("Content-Type", "application/gzip")
					w.Header().Set("Content-Disposition", "attachment; filename=test.tar.gz")
					_, err = io.Copy(w, &buf)
					assert.NoError(t, err)
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
			checkResult: func(t *testing.T, _ DownloadResponse, err error) {
				require.Error(t, err)
				require.Regexp(t, "^tar archive does not contain an mmdb file", err.Error())
			},
		},
	}

	ctx := context.Background()

	accountID := 10
	licenseKey := "license"

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			server := test.server(t)
			defer server.Close()

			c, err := New(
				accountID,
				licenseKey,
				WithEndpoint(server.URL),
			)
			require.NoError(t, err)

			res, err := c.Download(ctx, edition.EditionID, edition.MD5)
			test.checkResult(t, res, err)
		})
	}
}
