package geoipupdate

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/http2"

	"github.com/maxmind/geoipupdate/v7/client"
	"github.com/maxmind/geoipupdate/v7/internal"
	"github.com/maxmind/geoipupdate/v7/internal/geoipupdate/database"
)

// TestUpdaterOutput makes sure that the Updater outputs the result of its
// operation to stdout in json format.
func TestUpdaterOutput(t *testing.T) {
	now := time.Now().Truncate(time.Second).In(time.UTC)
	testTime := time.Date(2023, 4, 27, 12, 4, 48, 0, time.UTC)
	outputs := []client.DownloadResponse{
		{
			LastModified:    testTime,
			MD5:             "B",
			Reader:          io.NopCloser(strings.NewReader("")),
			UpdateAvailable: true,
		},
		{
			LastModified:    testTime,
			MD5:             "D",
			Reader:          io.NopCloser(strings.NewReader("")),
			UpdateAvailable: true,
		},
	}

	tempDir := t.TempDir()

	config := &Config{
		EditionIDs:  []string{"GeoLite2-City", "GeoLite2-Country"},
		LockFile:    filepath.Join(tempDir, ".geoipupdate.lock"),
		Output:      true,
		Parallelism: 1,
	}

	// capture the output of the `output` logger.
	logOutput := &bytes.Buffer{}

	// create a fake Updater with a mocked database reader and writer.
	u := &Updater{
		config:       config,
		output:       log.New(logOutput, "", 0),
		updateClient: &mockUpdateClient{i: 0, outputs: outputs},
		writer: &mockWriter{
			md5s: map[string]string{
				// These are the "MD5s" that we currently have before running an
				// update.
				"GeoLite2-City":    "A",
				"GeoLite2-Country": "C",
			},
		},
	}

	err := u.Run(context.Background())
	require.NoError(t, err)

	// make sure the expected output matches the input.
	var outputDatabases []database.ReadResult
	err = json.Unmarshal(logOutput.Bytes(), &outputDatabases)
	require.NoError(t, err)

	wantDatabases := []database.ReadResult{
		{
			EditionID:  "GeoLite2-City",
			OldHash:    "A",
			NewHash:    "B",
			ModifiedAt: testTime,
		},
		{
			EditionID:  "GeoLite2-Country",
			OldHash:    "C",
			NewHash:    "D",
			ModifiedAt: testTime,
		},
	}

	require.Equal(t, len(wantDatabases), len(outputDatabases))

	for i := 0; i < len(wantDatabases); i++ {
		require.Equal(t, wantDatabases[i].EditionID, outputDatabases[i].EditionID)
		require.Equal(t, wantDatabases[i].OldHash, outputDatabases[i].OldHash)
		require.Equal(t, wantDatabases[i].NewHash, outputDatabases[i].NewHash)
		require.Equal(t, wantDatabases[i].ModifiedAt, outputDatabases[i].ModifiedAt)
		// comparing time wasn't supported with require in older go versions.
		if !afterOrEqual(outputDatabases[i].CheckedAt, now) {
			t.Errorf("database %s was not updated", outputDatabases[i].EditionID)
		}
	}

	// Test with a write error.

	u.updateClient.(*mockUpdateClient).i = 0

	streamErr := http2.StreamError{
		Code: http2.ErrCodeInternal,
	}
	u.writer = &mockWriter{
		writeFunc: func(_ string, _ io.ReadCloser, _ string, _ time.Time) error {
			return streamErr
		},
	}

	err = u.Run(context.Background())
	require.ErrorIs(t, err, streamErr)
}

func TestRetryWhenWriting(t *testing.T) {
	tempDir := t.TempDir()

	databaseDir := filepath.Join(tempDir, "databases")

	// Create a databases folder
	err := os.MkdirAll(databaseDir, 0o750)
	require.NoError(t, err)

	try := 0
	sv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Mocking the metadata endpoint.
		if r.URL.Path == "/geoip/updates/metadata" {
			w.Header().Set("Content-Type", "application/json")

			// The md5 here bleongs to the tar.gz sent below.
			metadata := []byte(
				`{"databases":[{"edition_id":"foo-db-name",` +
					`"md5":"83e01ba43c2a66e30cb3007c1a300c78","date":"2023-04-27"}]}`)
			_, err := w.Write(metadata)
			require.NoError(t, err)

			return
		}

		w.Header().Set("Last-Modified", "Wed, 27 Apr 2023 12:04:48 GMT")

		gzWriter := gzip.NewWriter(w)
		defer gzWriter.Close()
		tarWriter := tar.NewWriter(gzWriter)
		defer tarWriter.Close()

		info := mockFileInfo{
			name: "foo-db-name.mmdb",
			size: 1000,
		}
		header, err := tar.FileInfoHeader(info, info.Name())
		require.NoError(t, err)
		header.Name = "foo-db-name.mmdb"

		// Create a tar Header from the FileInfo data
		err = tarWriter.WriteHeader(header)
		require.NoError(t, err)

		bytesToWrite := 1000
		if try == 0 {
			// In the first try, we create a bad tar.gz file.
			// That has less than the size defined in the header.
			bytesToWrite = 100
		}

		for i := 0; i < bytesToWrite; i++ {
			_, err = tarWriter.Write([]byte("t"))
			require.NoError(t, err)
		}
		try++
	}))
	defer sv.Close()

	config := &Config{
		AccountID:         10,
		URL:               sv.URL,
		EditionIDs:        []string{"foo-db-name"},
		LicenseKey:        "foo",
		LockFile:          filepath.Join(tempDir, ".geoipupdate.lock"),
		Output:            true,
		Parallelism:       1,
		RetryFor:          5 * time.Minute,
		DatabaseDirectory: databaseDir,
	}

	logOutput := &bytes.Buffer{}

	updateClient, err := client.New(
		config.AccountID,
		config.LicenseKey,
		client.WithEndpoint(config.URL),
	)
	require.NoError(t, err)

	writer, err := database.NewLocalFileWriter(
		config.DatabaseDirectory,
		config.PreserveFileTimes,
		config.Verbose,
	)
	require.NoError(t, err)

	u := &Updater{
		config:       config,
		output:       log.New(logOutput, "", 0),
		updateClient: updateClient,
		writer:       writer,
	}

	ctx := context.Background()

	jobProcessor := internal.NewJobProcessor(ctx, 1)
	processFunc := func(ctx context.Context) error {
		_, err = u.downloadEdition(
			ctx,
			"foo-db-name",
			u.updateClient,
			u.writer,
		)

		return err
	}
	jobProcessor.Add(processFunc)

	err = jobProcessor.Run(ctx)
	require.NoError(t, err)

	assert.Empty(t, logOutput.String())
}

type mockUpdateClient struct {
	i       int
	outputs []client.DownloadResponse
}

func (m *mockUpdateClient) Download(
	_ context.Context,
	_,
	_ string,
) (client.DownloadResponse, error) {
	if m.i >= len(m.outputs) {
		return client.DownloadResponse{}, errors.New("out of bounds")
	}
	res := m.outputs[m.i]
	m.i++
	return res, nil
}

type mockWriter struct {
	md5s      map[string]string
	writeFunc func(string, io.ReadCloser, string, time.Time) error
}

func (w *mockWriter) Write(
	editionID string,
	reader io.ReadCloser,
	md5 string,
	lastModified time.Time,
) error {
	if w.writeFunc != nil {
		return w.writeFunc(editionID, reader, md5, lastModified)
	}

	return nil
}

func (w mockWriter) GetHash(editionID string) (string, error) {
	return w.md5s[editionID], nil
}

func afterOrEqual(t1, t2 time.Time) bool {
	return t1.After(t2) || t1.Equal(t2)
}

type mockFileInfo struct {
	name string
	size int64
}

func (info mockFileInfo) Name() string {
	return info.name
}

func (info mockFileInfo) Size() int64 {
	return info.size
}

func (info mockFileInfo) IsDir() bool {
	return false
}

func (info mockFileInfo) Mode() os.FileMode {
	return 0
}

func (info mockFileInfo) ModTime() time.Time {
	return time.Now()
}

func (info mockFileInfo) Sys() any {
	return nil
}
