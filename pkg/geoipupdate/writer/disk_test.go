package writer

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/maxmind/geoipupdate/v6/pkg/geoipupdate/api"
	"github.com/stretchr/testify/require"
)

// TestGetHash checks the GetHash method of a diskWriter.
func TestGetHash(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "db")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	d := NewDiskWriter(tempDir, false)

	// returns a zero hash for a non existing edition.
	hash, err := d.GetHash("NewEdition")
	require.NoError(t, err)
	require.Equal(t, zeroMD5, hash)

	// returns the correct md5 for an existing edition.
	edition := "edition-1"
	dbFile := filepath.Join(tempDir, edition+extension)

	err = os.WriteFile(dbFile, []byte("edition-1 content"), os.ModePerm)
	require.NoError(t, err)

	hash, err = d.GetHash(edition)
	require.NoError(t, err)
	require.Equal(t, "618dd27a10de24809ec160d6807f363f", hash)
}

// TestDiskWriter checks the Write method of a diskWriter.
func TestDiskWriter(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "db")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	now := time.Now()

	edition := api.Metadata{
		EditionID: "edition-1",
		Date:      "2024-02-02",
		MD5:       "618dd27a10de24809ec160d6807f363f",
	}

	dbContent := "edition-1 content"
	tests := []struct {
		description      string
		preserveFileTime bool
		reader           io.Reader
		checkResult      func(t *testing.T, err error)
	}{
		{
			description:      "successful write",
			preserveFileTime: false,
			reader:           strings.NewReader(dbContent),
			checkResult: func(t *testing.T, err error) {
				require.NoError(t, err)

				dbFile := filepath.Join(tempDir, edition.EditionID+extension)

				//nolint:gosec // we need to read the content of the file in this test.
				fileContent, err := os.ReadFile(dbFile)
				require.NoError(t, err)
				require.Equal(t, dbContent, string(fileContent))

				database, err := os.Stat(dbFile)
				require.NoError(t, err)
				// acomodate time drift
				require.GreaterOrEqual(t, database.ModTime(), now.Add(-1*time.Hour))
			},
		},
		{
			description:      "successful write with db time preserved",
			preserveFileTime: true,
			reader:           strings.NewReader(dbContent),
			checkResult: func(t *testing.T, err error) {
				require.NoError(t, err)

				dbFile := filepath.Join(tempDir, edition.EditionID+extension)

				//nolint:gosec // we need to read the content of the file in this test.
				fileContent, err := os.ReadFile(dbFile)
				require.NoError(t, err)
				require.Equal(t, dbContent, string(fileContent))

				modTime, err := api.ParseTime(edition.Date)
				require.NoError(t, err)

				database, err := os.Stat(dbFile)
				require.NoError(t, err)
				require.Equal(t, modTime, database.ModTime())
			},
		},
		{
			description:      "file hash does not match metadata",
			preserveFileTime: true,
			reader:           strings.NewReader("malformed content"),
			checkResult: func(t *testing.T, err error) {
				require.Error(t, err)
				//nolint:lll
				require.Regexp(t, "^validating hash for edition-1: md5 of new database .* does not match expected md5", err.Error())
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			d := NewDiskWriter(tempDir, test.preserveFileTime)
			err = d.Write(edition, test.reader)
			test.checkResult(t, err)
		})
	}
}
