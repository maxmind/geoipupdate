package database

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestLocalFileWriterWrite tests functionality of the LocalFileWriter.Write method.
func TestLocalFileWriterWrite(t *testing.T) {
	testTime := time.Date(2023, 4, 10, 12, 47, 31, 0, time.UTC)

	tests := []struct {
		description string
		//nolint:revive // support older versions
		checkErr         func(require.TestingT, error, ...interface{})
		preserveFileTime bool
		//nolint:revive // support older versions
		checkTime func(require.TestingT, interface{}, interface{}, ...interface{})
		result    *ReadResult
	}{
		{
			description:      "success",
			checkErr:         require.NoError,
			preserveFileTime: true,
			checkTime:        require.Equal,
			result: &ReadResult{
				reader:     getReader(t, "database content"),
				EditionID:  "GeoIP2-City",
				OldHash:    "",
				NewHash:    "cfa36ddc8279b5483a5aa25e9a6151f4",
				ModifiedAt: testTime,
			},
		}, {
			description:      "hash does not match",
			checkErr:         require.Error,
			preserveFileTime: true,
			checkTime:        require.Equal,
			result: &ReadResult{
				reader:     getReader(t, "database content"),
				EditionID:  "GeoIP2-City",
				OldHash:    "",
				NewHash:    "badhash",
				ModifiedAt: testTime,
			},
		}, {
			description:      "hash case does not matter",
			checkErr:         require.NoError,
			preserveFileTime: true,
			checkTime:        require.Equal,
			result: &ReadResult{
				reader:     getReader(t, "database content"),
				EditionID:  "GeoIP2-City",
				OldHash:    "",
				NewHash:    "cfa36ddc8279b5483a5aa25e9a6151f4",
				ModifiedAt: testTime,
			},
		}, {
			description:      "do not preserve file modification time",
			checkErr:         require.NoError,
			preserveFileTime: false,
			checkTime:        require.NotEqual,
			result: &ReadResult{
				reader:     getReader(t, "database content"),
				EditionID:  "GeoIP2-City",
				OldHash:    "",
				NewHash:    "CFA36DDC8279B5483A5AA25E9A6151F4",
				ModifiedAt: testTime,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			tempDir := t.TempDir()
			defer test.result.reader.Close()

			fw, err := NewLocalFileWriter(tempDir, test.preserveFileTime, false)
			require.NoError(t, err)

			err = fw.Write(test.result)
			test.checkErr(t, err)
			if err == nil {
				database, err := os.Stat(fw.getFilePath(test.result.EditionID))
				require.NoError(t, err)

				test.checkTime(t, database.ModTime().UTC(), testTime)
			}
		})
	}
}

// TestLocalFileWriterGetHash tests functionality of the LocalFileWriter.GetHash method.
func TestLocalFileWriterGetHash(t *testing.T) {
	result := &ReadResult{
		reader:     getReader(t, "database content"),
		EditionID:  "GeoIP2-City",
		OldHash:    "",
		NewHash:    "cfa36ddc8279b5483a5aa25e9a6151f4",
		ModifiedAt: time.Time{},
	}

	tempDir := t.TempDir()

	defer result.reader.Close()

	fw, err := NewLocalFileWriter(tempDir, false, false)
	require.NoError(t, err)

	err = fw.Write(result)
	require.NoError(t, err)

	// returns the correct hash for an existing database.
	hash, err := fw.GetHash(result.EditionID)
	require.NoError(t, err)
	require.Equal(t, hash, result.NewHash)

	// returns a zero hash for a non existing edition.
	hash, err = fw.GetHash("NewEdition")
	require.NoError(t, err)
	require.Equal(t, ZeroMD5, hash)
}
