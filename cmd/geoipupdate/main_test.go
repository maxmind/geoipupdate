package main

import (
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateEdition(t *testing.T) {
	tests := []struct {
		Description     string
		CreateDirectory bool
		DatabaseBefore  string
		DatabaseAfter   string
		FilenameStatus  int
		FilenameBody    string
		DownloadStatus  int
		DownloadBody    string
		DownloadHeaders map[string]string
		ExpectedTime    time.Time
		Err             string
	}{
		{
			Description:     "Initial download, success",
			CreateDirectory: true,
			DatabaseAfter:   "database goes here",
			FilenameStatus:  http.StatusOK,
			FilenameBody:    "GeoIP2-City.mmdb",
			DownloadStatus:  http.StatusOK,
			DownloadBody:    "database goes here",
		},
		{
			Description:     "No update, success",
			CreateDirectory: true,
			DatabaseBefore:  "database goes here",
			DatabaseAfter:   "database goes here",
			FilenameStatus:  http.StatusOK,
			FilenameBody:    "GeoIP2-City.mmdb",
			DownloadStatus:  http.StatusNotModified,
			DownloadBody:    "database goes here",
		},
		{
			Description:     "Update, success",
			CreateDirectory: true,
			DatabaseBefore:  "database goes here",
			DatabaseAfter:   "new database goes here",
			FilenameStatus:  http.StatusOK,
			FilenameBody:    "GeoIP2-City.mmdb",
			DownloadStatus:  http.StatusOK,
			DownloadBody:    "new database goes here",
		},
		{
			Description:     "Update, success, and modification time is set",
			CreateDirectory: true,
			DatabaseBefore:  "new database goes here",
			DatabaseAfter:   "newer database goes here",
			FilenameStatus:  http.StatusOK,
			FilenameBody:    "GeoIP2-City.mmdb",
			DownloadStatus:  http.StatusOK,
			DownloadBody:    "newer database goes here",
			DownloadHeaders: map[string]string{
				"Last-Modified": time.Date(2018, 7, 24, 0, 0, 0, 0, time.UTC).Format(time.RFC1123),
			},
			ExpectedTime: time.Date(2018, 7, 24, 0, 0, 0, 0, time.UTC),
		},
		{
			Description:     "Get filename fails",
			CreateDirectory: true,
			DatabaseBefore:  "database goes here",
			DatabaseAfter:   "database goes here",
			FilenameStatus:  http.StatusBadRequest,
			Err:             "error retrieving filename: unexpected HTTP status code: 400 Bad Request",
		},
		{
			Description:     "Get filename is missing body",
			CreateDirectory: true,
			DatabaseBefore:  "database goes here",
			DatabaseAfter:   "database goes here",
			FilenameStatus:  http.StatusOK,
			Err:             "error retrieving filename: response body is empty",
		},
		{
			Description:     "Get filename has newlines",
			CreateDirectory: true,
			DatabaseBefore:  "database goes here",
			DatabaseAfter:   "database goes here",
			FilenameStatus:  http.StatusOK,
			FilenameBody:    "bad\nfilename",
			Err:             "error retrieving filename: invalid characters in filename",
		},
		{
			Description:     "Download request fails",
			CreateDirectory: true,
			DatabaseBefore:  "database goes here",
			DatabaseAfter:   "database goes here",
			FilenameStatus:  http.StatusOK,
			FilenameBody:    "GeoIP2-City.mmdb",
			DownloadStatus:  http.StatusBadRequest,
			Err:             "error updating: unexpected HTTP status code: 400 Bad Request",
		},
		{
			Description:     "Download request is missing X-Database-MD5",
			CreateDirectory: true,
			DatabaseBefore:  "database goes here",
			DatabaseAfter:   "database goes here",
			FilenameStatus:  http.StatusOK,
			FilenameBody:    "GeoIP2-City.mmdb",
			DownloadStatus:  http.StatusOK,
			DownloadBody:    "new database goes here",
			DownloadHeaders: map[string]string{
				"X-Database-MD5": "",
			},
			Err: "error updating: no X-Database-MD5 header found",
		},
		{
			Description:    "Download fails because database directory does not exist",
			FilenameStatus: http.StatusOK,
			FilenameBody:   "GeoIP2-City.mmdb",
			DownloadStatus: http.StatusOK,
			DownloadBody:   "new database goes here",
			Err:            `error updating: error creating file: open \S+GeoIP2-City\.mmdb\.test: (?:no such file or directory|The system cannot find the path specified)`,
		},
		{
			Description:     "Download fails because provided checksum does not match",
			CreateDirectory: true,
			DatabaseBefore:  "database goes here",
			DatabaseAfter:   "database goes here",
			FilenameStatus:  http.StatusOK,
			FilenameBody:    "GeoIP2-City.mmdb",
			DownloadStatus:  http.StatusOK,
			DownloadBody:    "new database goes here",
			DownloadHeaders: map[string]string{
				"X-Database-MD5": "5d41402abc4b2a76b9719d911017c592", // "hello"
			},
			Err: `error updating: MD5 of new database \(985ecf3d7959b146208b3dc0189b21a5\) does not match expected MD5 \(5d41402abc4b2a76b9719d911017c592\)`,
		},
		{
			Description:     "Download request redirects are followed",
			CreateDirectory: true,
			DatabaseBefore:  "database goes here",
			DatabaseAfter:   "database goes here",
			FilenameStatus:  http.StatusOK,
			FilenameBody:    "GeoIP2-City.mmdb",
			DownloadStatus:  http.StatusMovedPermanently,
			DownloadHeaders: map[string]string{
				"Location": "/go-here",
			},
		},
		{
			Description:     "MD5 sums are case insensitive",
			CreateDirectory: true,
			DatabaseBefore:  "database goes here",
			DatabaseAfter:   "new database goes here",
			FilenameStatus:  http.StatusOK,
			FilenameBody:    "GeoIP2-City.mmdb",
			DownloadStatus:  http.StatusOK,
			DownloadBody:    "new database goes here",
			DownloadHeaders: map[string]string{
				"X-Database-MD5": "985ECF3D7959B146208B3DC0189B21A5",
			},
		},
	}

	updateRE := regexp.MustCompile(`\A/geoip/databases/\S+/update\z`)

	tempDir, err := ioutil.TempDir("", "gutest-")
	require.NoError(t, err)
	err = os.RemoveAll(tempDir)
	require.NoError(t, err)

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

					if updateRE.MatchString(r.URL.Path) {
						buf := &bytes.Buffer{}
						gzWriter := gzip.NewWriter(buf)
						md5Writer := md5.New()
						multiWriter := io.MultiWriter(gzWriter, md5Writer)
						_, err := multiWriter.Write([]byte(test.DownloadBody))
						require.NoError(t, err)
						err = gzWriter.Close()
						require.NoError(t, err)

						rw.Header().Set(
							"X-Database-MD5",
							fmt.Sprintf("%x", md5Writer.Sum(nil)),
						)
						if test.DownloadStatus == http.StatusOK {
							rw.Header().Set(
								"Last-Modified",
								time.Now().Format(time.RFC1123),
							)
						}
						for k, v := range test.DownloadHeaders {
							rw.Header().Set(k, v)
						}

						rw.WriteHeader(test.DownloadStatus)

						if test.DownloadStatus == http.StatusOK {
							_, err := rw.Write(buf.Bytes())
							require.NoError(t, err)
						}

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
		verbose := false
		if !test.ExpectedTime.IsZero() {
			config.PreserveFileTimes = true
		}

		if test.CreateDirectory {
			err := os.Mkdir(config.DatabaseDirectory, 0755)
			require.NoError(t, err)
		}

		currentDatabasePath := filepath.Join(
			config.DatabaseDirectory,
			"GeoIP2-City.mmdb",
		)
		if test.DatabaseBefore != "" {
			err := ioutil.WriteFile(
				currentDatabasePath,
				[]byte(test.DatabaseBefore),
				0644,
			)
			require.NoError(t, err)
		}

		err := updateEdition(config, verbose, config.EditionIDs[0])
		if test.Err == "" {
			assert.NoError(t, err, test.Description)
		} else {
			// regex because some errors have filenames.
			assert.Regexp(t, test.Err, err.Error(), test.Description)
		}

		server.Close()

		if test.DatabaseAfter != "" {
			buf, err := ioutil.ReadFile(currentDatabasePath)
			require.NoError(t, err, test.Description)
			assert.Equal(t, test.DatabaseAfter, string(buf))
		}

		if !test.ExpectedTime.IsZero() {
			fi, err := os.Stat(currentDatabasePath)
			require.NoError(t, err)
			assert.WithinDuration(t, test.ExpectedTime, fi.ModTime(), 0)
		}

		if test.CreateDirectory {
			err := os.RemoveAll(config.DatabaseDirectory)
			require.NoError(t, err)
		}
	}
}

func TestGetCurrentMD5(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "gutest-")
	require.NoError(t, err)
	defer func() {
		err := os.RemoveAll(tempDir)
		require.NoError(t, err)
	}()

	config := &Config{
		DatabaseDirectory: tempDir,
	}

	dirFile := filepath.Join(tempDir, "mydir")
	err = os.Mkdir(dirFile, 0755)
	require.NoError(t, err)

	verbose := false
	md5, err := getCurrentMD5(config, verbose, "mydir")
	assert.EqualError(t, err, "not a regular file")
	assert.Equal(t, "", md5)
}
