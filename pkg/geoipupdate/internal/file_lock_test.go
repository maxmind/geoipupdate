package internal

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestAcquireFileLock tests that a lock can be acquired multile times
// within a same process.
func TestAcquireFileLock(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gutest-")
	require.NoError(t, err)
	defer func() {
		err = os.RemoveAll(tempDir)
		require.NoError(t, err)
	}()

	fl, err := NewFileLock(filepath.Join(tempDir, ".geoipupdate.lock"), false)
	require.NoError(t, err)
	defer func() {
		err := fl.Release()
		require.NoError(t, err)
	}()

	// acquire lock
	err = fl.Acquire()
	require.NoError(t, err)
	require.True(t, fl.lock.Locked())

	// acquiring lock a second time within the same process
	// should succeed
	err = fl.Acquire()
	require.NoError(t, err)
	require.True(t, fl.lock.Locked())

	// release lock
	err = fl.Release()
	require.NoError(t, err)
	require.False(t, fl.lock.Locked())

	// acquire a released lock
	err = fl.Acquire()
	require.NoError(t, err)
	require.True(t, fl.lock.Locked())
}
