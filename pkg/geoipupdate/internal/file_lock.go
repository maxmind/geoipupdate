package internal

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/gofrs/flock"
)

// FileLock provides a file lock mechanism based on flock.
type FileLock struct {
	lock    *flock.Flock
	verbose bool
}

// NewFileLock creates a new instance of FileLock.
func NewFileLock(path string, verbose bool) (*FileLock, error) {
	err := os.MkdirAll(filepath.Dir(path), 0o750)
	if err != nil {
		return nil, fmt.Errorf("creating lock file directory: %w", err)
	}

	if verbose {
		log.Printf("Initializing file lock at %s", path)
	}

	return &FileLock{
		lock:    flock.New(path),
		verbose: verbose,
	}, nil
}

// Release unlocks the file lock.
func (f *FileLock) Release() error {
	if err := f.lock.Unlock(); err != nil {
		return fmt.Errorf("releasing file lock at %s: %w", f.lock.Path(), err)
	}
	if f.verbose {
		log.Printf("Lock file %s successfully released", f.lock.Path())
	}
	return nil
}

// Acquire tries to acquire a file lock.
// It is possible for multiple goroutines within the same process
// to acquire the same lock, so acquireLock is not thread safe in
// that sense, but protects access across different processes.
func (f *FileLock) Acquire() error {
	ok, err := f.lock.TryLock()
	if err != nil {
		return fmt.Errorf("acquiring file lock at %s: %w", f.lock.Path(), err)
	}
	if !ok {
		return fmt.Errorf("lock %s already acquired by another process", f.lock.Path())
	}
	if f.verbose {
		log.Printf("Acquired lock file at %s", f.lock.Path())
	}
	return nil
}
