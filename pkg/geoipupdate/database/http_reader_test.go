package database

import (
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestHTTPReader tests the functionality of the HTTPReader.Read method.
func TestHTTPReader(t *testing.T) {
	testTime := time.Date(2023, 4, 10, 12, 47, 31, 0, time.UTC)

	tests := []struct {
		description    string
		checkErr       func(require.TestingT, error, ...interface{}) //nolint:revive // support older versions
		requestEdition string
		requestHash    string
		responseStatus int
		responseBody   string
		responseHash   string
		responseTime   string
		result         *ReadResult
	}{
		{
			description:    "success",
			checkErr:       require.NoError,
			requestEdition: "GeoIP2-City",
			requestHash:    "fbe1786bfd80e1db9dc42ddaff868f38",
			responseStatus: http.StatusOK,
			responseBody:   "database content",
			responseHash:   "cfa36ddc8279b5483a5aa25e9a6151f4",
			responseTime:   testTime.Format(time.RFC1123),
			result: &ReadResult{
				reader:     getReader(t, "database content"),
				EditionID:  "GeoIP2-City",
				OldHash:    "fbe1786bfd80e1db9dc42ddaff868f38",
				NewHash:    "cfa36ddc8279b5483a5aa25e9a6151f4",
				ModifiedAt: testTime,
			},
		}, {
			description:    "no new update",
			checkErr:       require.NoError,
			requestEdition: "GeoIP2-City",
			requestHash:    "fbe1786bfd80e1db9dc42ddaff868f38",
			responseStatus: http.StatusNotModified,
			responseBody:   "",
			responseHash:   "",
			responseTime:   "",
			result: &ReadResult{
				reader:     nil,
				EditionID:  "GeoIP2-City",
				OldHash:    "fbe1786bfd80e1db9dc42ddaff868f38",
				NewHash:    "fbe1786bfd80e1db9dc42ddaff868f38",
				ModifiedAt: time.Time{},
			},
		}, {
			description:    "bad request",
			checkErr:       require.Error,
			requestEdition: "GeoIP2-City",
			requestHash:    "fbe1786bfd80e1db9dc42ddaff868f38",
			responseStatus: http.StatusBadRequest,
			responseBody:   "",
			responseHash:   "",
			responseTime:   "",
		}, {
			description:    "missing hash header",
			checkErr:       require.Error,
			requestEdition: "GeoIP2-City",
			requestHash:    "fbe1786bfd80e1db9dc42ddaff868f38",
			responseStatus: http.StatusOK,
			responseBody:   "database content",
			responseHash:   "",
			responseTime:   testTime.Format(time.RFC1123),
		}, {
			description:    "modified time header wrong format",
			checkErr:       require.Error,
			requestEdition: "GeoIP2-City",
			requestHash:    "fbe1786bfd80e1db9dc42ddaff868f38",
			responseStatus: http.StatusOK,
			responseBody:   "database content",
			responseHash:   "fbe1786bfd80e1db9dc42ddaff868f38",
			responseTime:   testTime.Format(time.Kitchen),
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			server := httptest.NewServer(
				http.HandlerFunc(
					func(w http.ResponseWriter, r *http.Request) {
						if test.responseStatus != http.StatusOK {
							w.WriteHeader(test.responseStatus)
							return
						}

						w.Header().Set("X-Database-MD5", test.responseHash)
						w.Header().Set("Last-Modified", test.responseTime)

						buf := &bytes.Buffer{}
						gzWriter := gzip.NewWriter(buf)
						_, err := gzWriter.Write([]byte(test.responseBody))
						require.NoError(t, err)
						require.NoError(t, gzWriter.Flush())
						require.NoError(t, gzWriter.Close())
						_, err = w.Write(buf.Bytes())
						require.NoError(t, err)
					},
				),
			)
			defer server.Close()

			reader := NewHTTPReader(
				nil,        // request proxy.
				server.URL, // fixed, as the server is mocked above.
				10,         // fixed, as it's not valuable for the purpose of the test.
				"license",  // fixed, as it's not valuable for the purpose of the test.
				0,          // zero means no retries.
				false,      // verbose
			)

			result, err := reader.Read(context.Background(), test.requestEdition, test.requestHash)
			test.checkErr(t, err)
			if err == nil {
				require.Equal(t, result.EditionID, test.result.EditionID)
				require.Equal(t, result.OldHash, test.result.OldHash)
				require.Equal(t, result.NewHash, test.result.NewHash)
				require.Equal(t, result.ModifiedAt, test.result.ModifiedAt)

				if test.result.reader != nil && result.reader != nil {
					defer result.reader.Close()
					defer test.result.reader.Close()
					resultDatabase, err := io.ReadAll(test.result.reader)
					require.NoError(t, err)
					expectedDatabase, err := io.ReadAll(result.reader)
					require.NoError(t, err)
					require.Equal(t, expectedDatabase, resultDatabase)
				}
			}
		})
	}
}

//nolint:unparam // complains that it always receives the same string to encode. ridiculous.
func getReader(t *testing.T, s string) io.ReadCloser {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	_, err := gz.Write([]byte(s))
	require.NoError(t, err)
	require.NoError(t, gz.Close())
	require.NoError(t, gz.Flush())
	r, err := gzip.NewReader(&buf)
	require.NoError(t, err)
	return r
}
