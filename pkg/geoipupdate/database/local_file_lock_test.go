package database

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCreateFileLock tests the initialization of the FileLock struct.
func TestCreateFileLock(t *testing.T) {
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
			ExpectedError: `database directory does not exist`,
		},
	}

	for _, test := range tests {
		t.Run(test.Description, func(t *testing.T) {
			tempDir, err := ioutil.TempDir("", "gutest-")
			require.NoError(t, err)
			defer func() {
				err = os.RemoveAll(tempDir)
				require.NoError(t, err)
			}()
			fl, err := NewFileLock(filepath.Join(tempDir, test.LockFilename), false)
			if err != nil {
				// regex because some errors have filenames.
				assert.Regexp(t, test.ExpectedError, err.Error(), test.Description)
			} else {
				defer func() {
					err := fl.Close()
					require.NoError(t, err)
				}()
				require.Equal(t, test.ExpectedError, "", test.Description)
			}
		})
	}
}

// TestAcquireFileLock tests that a lock can be acquired multile times
// within a same process.
func TestAcquireFileLock(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "gutest-")
	require.NoError(t, err)
	defer func() {
		err = os.RemoveAll(tempDir)
		require.NoError(t, err)
	}()

	fl, err := NewFileLock(filepath.Join(tempDir, ".geoipupdate.lock"), false)
	require.NoError(t, err)
	defer func() {
		err := fl.Close()
		require.NoError(t, err)
	}()

	lock, err := fl.acquireLock()
	require.NoError(t, err)
	require.True(t, lock.Locked())

	lock2, err := fl.acquireLock()
	require.NoError(t, err)
	require.True(t, lock2.Locked())
}
