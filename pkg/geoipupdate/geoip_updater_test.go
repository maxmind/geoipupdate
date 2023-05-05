package geoipupdate

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"path/filepath"
	"testing"
	"time"

	"github.com/maxmind/geoipupdate/v5/pkg/geoipupdate/database"
	"github.com/stretchr/testify/require"
)

// TestClientOutput makes sure that the client outputs the result of it's
// operation to stdout in json format.
func TestClientOutput(t *testing.T) {
	now := time.Now().Truncate(time.Second).In(time.UTC)
	testTime := time.Date(2023, 4, 27, 12, 4, 48, 0, time.UTC)
	databases := []database.ReadResult{
		{
			EditionID:  "GeoLite2-City",
			OldHash:    "A",
			NewHash:    "B",
			ModifiedAt: testTime,
		}, {
			EditionID:  "GeoIP2-Country",
			OldHash:    "C",
			NewHash:    "D",
			ModifiedAt: testTime,
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

	// create a fake client with a mocked database reader and writer.
	c := &Client{
		config: config,
		getReader: func() (database.Reader, error) {
			return &mockReader{i: 0, result: databases}, nil
		},
		getWriter: func() (database.Writer, error) {
			return &mockWriter{}, nil
		},
		output: log.New(logOutput, "", 0),
	}

	// run the client
	err := c.Run(context.Background())
	require.NoError(t, err)

	// make sure the expected output matches the input.
	var outputDatabases []database.ReadResult
	err = json.Unmarshal(logOutput.Bytes(), &outputDatabases)
	require.NoError(t, err)
	require.Equal(t, len(outputDatabases), len(databases))

	for i := 0; i < len(databases); i++ {
		require.Equal(t, databases[i].EditionID, outputDatabases[i].EditionID)
		require.Equal(t, databases[i].OldHash, outputDatabases[i].OldHash)
		require.Equal(t, databases[i].NewHash, outputDatabases[i].NewHash)
		require.Equal(t, databases[i].ModifiedAt, outputDatabases[i].ModifiedAt)
		// comparing time wasn't supported with require in older go versions.
		if !afterOrEqual(outputDatabases[i].CheckedAt, now) {
			t.Errorf("database %s was not updated", outputDatabases[i].EditionID)
		}
	}
}

type mockReader struct {
	i      int
	result []database.ReadResult
}

func (mr *mockReader) Read(_ context.Context, _, _ string) (*database.ReadResult, error) {
	if mr.i >= len(mr.result) {
		return nil, errors.New("out of bounds")
	}
	res := mr.result[mr.i]
	mr.i++
	return &res, nil
}

type mockWriter struct{}

func (w *mockWriter) Write(_ *database.ReadResult) error { return nil }
func (w mockWriter) GetHash(_ string) (string, error)    { return "", nil }

func afterOrEqual(t1, t2 time.Time) bool {
	return t1.After(t2) || t1.Equal(t2)
}
