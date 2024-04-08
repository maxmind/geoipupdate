package client

import (
	"context"
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

			result, err := c.getMetadata(ctx, "edition-1")
			test.checkResult(t, result, err)
		})
	}
}
