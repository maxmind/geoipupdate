package database

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/gofrs/flock"
)

// FileLock provides a mechanism for blocking multiple instances of geoipupdate
// from downloading databases using the same lock file.
type FileLock struct {
	lock    *flock.Flock
	verbose bool
}

// NewFileLock creates a new instance of FileLock.
func NewFileLock(lockFilePath string, verbose bool) (*FileLock, error) {
	databaseDir, err := os.Stat(filepath.Dir(lockFilePath))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("database directory does not exist: %w", err)
		} else {
			return nil, fmt.Errorf("error checking database directory: %w", err)
		}
	}

	if !databaseDir.IsDir() {
		return nil, errors.New("database directory is not a directory")
	}

	if verbose {
		log.Printf("Initializing lock at %s", lockFilePath)
	}

	return &FileLock{
		lock:    flock.New(lockFilePath),
		verbose: verbose,
	}, nil
}

// Close unlocks the file lock.
func (f *FileLock) Close() error {
	if err := f.lock.Unlock(); err != nil {
		return fmt.Errorf("error releasing lock file: %w", err)
	}
	return nil
}

// acquireLock tries to acquire the file lock.
// It is possible for multiple goroutines within the same process
// to aquire the same lock, so aquireLock is not thread safe in
// that sense, but protects access accross different processes.
func (f *FileLock) acquireLock() (*flock.Flock, error) {
	ok, err := f.lock.TryLock()
	if err != nil {
		return nil, fmt.Errorf("error acquiring lock %s: %w", f.lock.Path(), err)
	}
	if !ok {
		return nil, fmt.Errorf("lock %s already acquired by another process", f.lock.Path())
	}
	if f.verbose {
		log.Printf("Acquired lock file %s", f.lock.Path())
	}
	return f.lock, nil
}
