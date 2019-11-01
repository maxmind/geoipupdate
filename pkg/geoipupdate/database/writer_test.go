package database

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestCreateLockFile(t *testing.T) {
	tests := []struct {
		Description   string
		LockFilename  string
		ExpectedError string
	}{
		{
			Description:   "Database shouldn't have errors with good file paths",
			LockFilename:  ".geoipupdate.lock",
			ExpectedError: "",
		},
		{
			Description:   "Database should fail to build with bad file path",
			LockFilename:  "bad/file/path.geoipupdate.lock",
			ExpectedError: `database directory is not available.*no such file or directory`,
		},
	}

	for _, test := range tests {
		tempDir, err := ioutil.TempDir("", "gutest-")
		require.NoError(t, err)
		err = os.RemoveAll(tempDir)
		require.NoError(t, err)
		t.Run(test.Description, func(t *testing.T) {
			_, err := CreateLockFile(filepath.Join(tempDir, test.LockFilename), false)
			if err != nil {
				// regex because some errors have filenames.
				assert.Regexp(t, test.ExpectedError, err.Error(), test.Description)
			} else {
				require.Equal(t, test.ExpectedError, "", test.Description)
			}
		})
	}
}
