package lock

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestAcquireFileLock tests that a lock can be acquired multile times
// within a same process.
func TestAcquireFileLock(t *testing.T) {
	tempDir := t.TempDir()

	fl, err := NewFileLock(filepath.Join(tempDir, ".geoipupdate.lock"))
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
