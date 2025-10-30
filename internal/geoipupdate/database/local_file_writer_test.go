package database

import (
	"io"
	"os"
	"strings"
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
		checkErr         func(require.TestingT, error, ...any)
		preserveFileTime bool
		//nolint:revive // support older versions
		checkTime    func(require.TestingT, any, any, ...any)
		editionID    string
		reader       io.ReadCloser
		newMD5       string
		lastModified time.Time
	}{
		{
			description:      "success",
			checkErr:         require.NoError,
			preserveFileTime: true,
			checkTime:        require.Equal,
			editionID:        "GeoIP2-City",
			reader:           io.NopCloser(strings.NewReader("database content")),
			newMD5:           "cfa36ddc8279b5483a5aa25e9a6151f4",
			lastModified:     testTime,
		}, {
			description:      "hash does not match",
			checkErr:         require.Error,
			preserveFileTime: true,
			checkTime:        require.Equal,
			editionID:        "GeoIP2-City",
			reader:           io.NopCloser(strings.NewReader("database content")),
			newMD5:           "badhash",
			lastModified:     testTime,
		}, {
			description:      "hash case does not matter",
			checkErr:         require.NoError,
			preserveFileTime: true,
			checkTime:        require.Equal,
			editionID:        "GeoIP2-City",
			reader:           io.NopCloser(strings.NewReader("database content")),
			newMD5:           "cfa36ddc8279b5483a5aa25e9a6151f4",
			lastModified:     testTime,
		}, {
			description:      "do not preserve file modification time",
			checkErr:         require.NoError,
			preserveFileTime: false,
			checkTime:        require.NotEqual,
			editionID:        "GeoIP2-City",
			reader:           io.NopCloser(strings.NewReader("database content")),
			newMD5:           "CFA36DDC8279B5483A5AA25E9A6151F4",
			lastModified:     testTime,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			tempDir := t.TempDir()

			fw, err := NewLocalFileWriter(tempDir, test.preserveFileTime, false)
			require.NoError(t, err)

			err = fw.Write(
				test.editionID,
				test.reader,
				test.newMD5,
				test.lastModified,
			)
			test.checkErr(t, err)
			if err == nil {
				database, err := os.Stat(fw.getFilePath(test.editionID))
				require.NoError(t, err)

				test.checkTime(t, database.ModTime().UTC(), testTime)
			}
		})
	}
}

// TestLocalFileWriterGetHash tests functionality of the LocalFileWriter.GetHash method.
func TestLocalFileWriterGetHash(t *testing.T) {
	editionID := "GeoIP2-City"
	reader := io.NopCloser(strings.NewReader("database content"))
	newMD5 := "cfa36ddc8279b5483a5aa25e9a6151f4"
	lastModified := time.Time{}

	tempDir := t.TempDir()

	fw, err := NewLocalFileWriter(tempDir, false, false)
	require.NoError(t, err)

	err = fw.Write(editionID, reader, newMD5, lastModified)
	require.NoError(t, err)

	// returns the correct hash for an existing database.
	hash, err := fw.GetHash(editionID)
	require.NoError(t, err)
	require.Equal(t, hash, newMD5)

	// returns a zero hash for a non existing edition.
	hash, err = fw.GetHash("NewEdition")
	require.NoError(t, err)
	require.Equal(t, ZeroMD5, hash)
}
