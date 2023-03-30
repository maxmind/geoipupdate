package main

import (
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/maxmind/geoipupdate/v4/pkg/geoipupdate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMultipleDatabaseDownload(t *testing.T) {
	databaseContent := "database content goes here"

	server := httptest.NewServer(
		http.HandlerFunc(
			func(rw http.ResponseWriter, r *http.Request) {
				err := r.ParseForm()
				require.NoError(t, err, "parse form")

				if strings.HasPrefix(r.URL.Path, "/geoip/databases") {
					buf := &bytes.Buffer{}
					gzWriter := gzip.NewWriter(buf)
					md5Writer := md5.New()
					multiWriter := io.MultiWriter(gzWriter, md5Writer)
					_, err := multiWriter.Write([]byte(
						databaseContent + " " + r.URL.Path,
					))
					require.NoError(t, err)
					err = gzWriter.Close()
					require.NoError(t, err)

					rw.Header().Set(
						"X-Database-MD5",
						fmt.Sprintf("%x", md5Writer.Sum(nil)),
					)
					rw.Header().Set("Last-Modified", time.Now().Format(time.RFC1123))

					_, err = rw.Write(buf.Bytes())
					require.NoError(t, err)

					return
				}

				rw.WriteHeader(http.StatusBadRequest)
			},
		),
	)
	defer server.Close()

	client := server.Client()

	tempDir, err := ioutil.TempDir("", "gutest-")
	require.NoError(t, err)
	defer func() {
		err := os.RemoveAll(tempDir)
		require.NoError(t, err)
	}()

	config := &geoipupdate.Config{
		AccountID:         123,
		DatabaseDirectory: tempDir,
		EditionIDs:        []string{"GeoLite2-City", "GeoLite2-Country"},
		LicenseKey:        "testing",
		LockFile:          filepath.Join(tempDir, ".geoipupdate.lock"),
		URL:               server.URL,
		Parallelism:       1,
	}

	logOutput := &bytes.Buffer{}
	log.SetOutput(logOutput)

	downloadFunc := func(editionID string) error {
		return download(client, config, editionID)
	}

	err = run(config, downloadFunc)
	assert.NoError(t, err, "run successfully")

	assert.Equal(t, "", logOutput.String(), "no logged output")

	for _, editionID := range config.EditionIDs {
		path := filepath.Join(config.DatabaseDirectory, editionID+".mmdb")
		buf, err := ioutil.ReadFile(filepath.Clean(path))
		require.NoError(t, err, "read file")
		assert.Equal(
			t,
			databaseContent+" /geoip/databases/"+editionID+"/update",
			string(buf),
			"correct database",
		)
	}
}

// TestParallelDatabaseDownload tests the parallel database download functionality
// and ensures that the maximum number of allowed goroutines does not exceed the value
// set in the config.
func TestParallelDatabaseDownload(t *testing.T) {
	simulatedDownloadDuration := 5 * time.Millisecond
	editions := []string{"one", "two", "three", "four", "five", "six", "seven", "eight", "nine", "ten"}

	tests := []struct {
		Description       string
		Parallelism       int
		durationCheckFunc func(elapsed time.Duration) bool
	}{{
		Description: "sequential downloads",
		Parallelism: 1,
		durationCheckFunc: func(elapsed time.Duration) bool {
			return elapsed >= time.Duration(len(editions))*simulatedDownloadDuration
		},
	}, {
		Description: "parallel downloads",
		Parallelism: 3,
		durationCheckFunc: func(elapsed time.Duration) bool {
			// all downloads should execute in a duration that is
			// less than what it would have taken to execute them
			// sequentially
			return elapsed < time.Duration(len(editions))*simulatedDownloadDuration
		},
	}}

	for _, test := range tests {
		t.Run(test.Description, func(t *testing.T) {
			config := &geoipupdate.Config{
				EditionIDs:  editions,
				Parallelism: test.Parallelism,
			}

			doneCh := make(chan struct{})

			var lock sync.Mutex
			runningGoroutines := 0
			maxConcurrentGoroutines := 0

			// A mock download function that is used to gather data
			// about the number of goroutines called.
			downloadFunc := func(s string) error {
				lock.Lock()
				runningGoroutines++
				if runningGoroutines > maxConcurrentGoroutines {
					maxConcurrentGoroutines = runningGoroutines
				}
				lock.Unlock()

				time.Sleep(simulatedDownloadDuration)

				lock.Lock()
				runningGoroutines--
				lock.Unlock()
				return nil
			}

			// Execute run in a goroutine so that we can exit early if the test
			// hangs or takes too long to execute.
			start := time.Now()
			var elapsed time.Duration
			go func() {
				err := run(config, downloadFunc)
				assert.NoError(t, err, "run executed successfully")
				close(doneCh)
			}()

			// Wait for run to complete or timeout after a certain duration
			select {
			case <-doneCh:
				elapsed = time.Since(start)
			case <-time.After(100 * time.Millisecond):
				t.Errorf("Timeout waiting for function completion")
			}

			// The maximum number of parallel downloads executed should not exceed
			// the number defined in the configuration.
			if maxConcurrentGoroutines > config.Parallelism {
				t.Errorf("Expected %d concurrent download processes, but got %d", config.Parallelism, maxConcurrentGoroutines)
			}

			if !test.durationCheckFunc(elapsed) {
				t.Errorf("downloads completed in %+v", elapsed)
			}
		})
	}
}
