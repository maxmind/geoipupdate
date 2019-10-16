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
	tempDir, err := ioutil.TempDir("", "gutest-")
	require.NoError(t, err)
	defer func() {
		err = os.RemoveAll(tempDir)
		require.NoError(t, err)
	}()

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
			FilePath:      "bad/file/path/GeoIP2-City.mmdb",
			LockFilePath:  ".geoipupdate.lock",
			ExpectedError: `Encountered an error creating file.*no such file or directory`,
		},
		{
			Description:   "Database should fail to build with bad file path",
			FilePath:      "GeoIP2-City.mmdb",
			LockFilePath:  "bad/file/path.geoipupdate.lock",
			ExpectedError: `error acquiring a lock.*no such file or directory`,
		},
	}

	for _, test := range tests {
		_, err := NewLocalFileDatabaseWriter(filepath.Join(tempDir, test.FilePath),
			filepath.Join(tempDir, test.LockFilePath), false)
		if err != nil {
			// regex because some errors have filenames.
			assert.Regexp(t, test.ExpectedError, err.Error(), test.Description)
		} else {
			require.Equal(t, test.ExpectedError, "", test.Description)
		}
	}

}
