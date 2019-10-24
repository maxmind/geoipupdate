package database

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestNewDatabaseWriter(t *testing.T) {

	tests := []struct {
		Description   string
		FilePath      string
		LockFilePath  string
		ExpectedError string
	}{
		{
			Description:   "Database shouldn't have errors with good file paths",
			FilePath:      "GeoIP2-City.mmdb",
			LockFilePath:  ".geoipupdate.lock",
			ExpectedError: "",
		},
		{
			Description:   "Database should fail to build with bad file path",
			FilePath:      "GeoIP2-City.mmdb",
			LockFilePath:  "bad/file/path.geoipupdate.lock",
			ExpectedError: `database directory is not available`,
		},
	}

	for _, test := range tests {
		tempDir, err := ioutil.TempDir("", "gutest-")
		require.NoError(t, err)
		err = os.RemoveAll(tempDir)
		require.NoError(t, err)
		t.Run(test.Description, func(t *testing.T) {
			_, err = NewLocalFileDatabaseWriter(filepath.Join(tempDir, test.FilePath),
				filepath.Join(tempDir, test.LockFilePath), false)
			if err != nil {
				// regex because some errors have filenames.
				assert.Regexp(t, test.ExpectedError, err.Error(), test.Description)
			} else {
				require.Equal(t, test.ExpectedError, "", test.Description)
			}
		})
	}

}
