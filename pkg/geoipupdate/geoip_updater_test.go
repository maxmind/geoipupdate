package geoipupdate

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestGetFileName(t *testing.T) {

	tests := []struct {
		Description    string
		FilenameStatus int
		FilenameBody   string
		ExpectedError  string
		ExpectedOutput string
	}{
		{
			Description:    "Simple Success",
			FilenameStatus: http.StatusOK,
			FilenameBody:   "aSimpleFileName",
			ExpectedOutput: "aSimpleFileName",
		},
		{
			Description:    "Get filename fails",
			FilenameStatus: http.StatusBadRequest,
			ExpectedError:  "unexpected HTTP status code: 400 Bad Request: ",
		},
		{
			Description:    "Get filename is missing body",
			FilenameStatus: http.StatusOK,
			ExpectedError:  "response body is empty",
		},
		{
			Description:    "Get filename has newlines",
			FilenameStatus: http.StatusOK,
			FilenameBody:   "bad\nfilename",
			ExpectedError:  "invalid characters in filename",
		},
	}

	tempDir, err := ioutil.TempDir("", "gutest-")
	require.NoError(t, err)
	defer func() {
		err = os.RemoveAll(tempDir)
		require.NoError(t, err)
	}()

	for _, test := range tests {
		server := httptest.NewServer(
			http.HandlerFunc(
				func(rw http.ResponseWriter, r *http.Request) {
					if r.URL.Path == "/app/update_getfilename" {
						rw.WriteHeader(test.FilenameStatus)
						_, err := rw.Write([]byte(test.FilenameBody))
						require.NoError(t, err)
						return
					}
					if r.URL.Path == "/go-here" {
						rw.WriteHeader(http.StatusNotModified)
						return
					}
					rw.WriteHeader(http.StatusBadRequest)
				},
			),
		)

		config := &Config{
			AccountID:         123,
			DatabaseDirectory: tempDir,
			EditionIDs:        []string{"GeoIP2-City"},
			LicenseKey:        "testing",
			LockFile:          filepath.Join(tempDir, ".geoipupdate.lock"),
			URL:               server.URL,
		}
		client := NewClient(config)
		t.Run(test.Description, func(t *testing.T) {

			actualOutput, actualError := GetFilename(config, config.EditionIDs[0], client)

			assert.Equal(t, test.ExpectedOutput, actualOutput, test.Description)
			if test.ExpectedError != "" {
				require.Error(t, actualError, test.Description)
				assert.Equal(t, test.ExpectedError, actualError.Error(), test.Description)
			}
		})
	}

}
