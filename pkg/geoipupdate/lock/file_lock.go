package lock

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gofrs/flock"
)

// fileLock provides a file lock mechanism based on flock.
type fileLock struct {
	lock *flock.Flock
}

// NewFileLock creates a new instance of FileLock.
//
//nolint:revive // unexported type fileLock is not meant to be used as a standalone type.
func NewFileLock(path string) (*fileLock, error) {
	err := os.MkdirAll(filepath.Dir(path), 0o750)
	if err != nil {
		return nil, fmt.Errorf("creating lock file directory: %w", err)
	}

	return &fileLock{
		lock: flock.New(path),
	}, nil
}

// Release unlocks the file lock.
func (f *fileLock) Release() error {
	if err := f.lock.Unlock(); err != nil {
		return fmt.Errorf("releasing file lock at %s: %w", f.lock.Path(), err)
	}
	return nil
}

// Acquire tries to acquire a file lock.
// It is possible for multiple goroutines within the same process
// to acquire the same lock, so acquireLock is not thread safe in
// that sense, but protects access across different processes.
func (f *fileLock) Acquire() error {
	ok, err := f.lock.TryLock()
	if err != nil {
		return fmt.Errorf("acquiring file lock at %s: %w", f.lock.Path(), err)
	}
	if !ok {
		return fmt.Errorf("lock %s already acquired by another process", f.lock.Path())
	}
	return nil
}
