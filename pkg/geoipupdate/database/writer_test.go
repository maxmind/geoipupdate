package database

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			ExpectedError: `database directory is not available`,
		},
	}

	for _, test := range tests {
		t.Run(test.Description, func(t *testing.T) {
			tempDir, err := ioutil.TempDir("", "gutest-")
			require.NoError(t, err)
			err = os.RemoveAll(tempDir)
			require.NoError(t, err)
			_, err = CreateLockFile(filepath.Join(tempDir, test.LockFilename), false)
			if err != nil {
				// regex because some errors have filenames.
				assert.Regexp(t, test.ExpectedError, err.Error(), test.Description)
			} else {
				require.Equal(t, test.ExpectedError, "", test.Description)
			}
		})
	}
}
