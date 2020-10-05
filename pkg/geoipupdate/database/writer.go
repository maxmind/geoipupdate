package database

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/gofrs/flock"
	"github.com/pkg/errors"
)

// ZeroMD5 is the default value provided as an MD5 hash for a non-existent
// database.
const ZeroMD5 = "00000000000000000000000000000000"

// Writer provides an interface for writing a database to a target location.
type Writer interface {
	io.WriteCloser
	ValidHash(expectedHash string) error
	GetHash() string
	SetFileModificationTime(lastModified time.Time) error
	Commit() error
}

// CreateLockFile takes the provided filePath and lockFilePath name to create a
// file lock. All output errors are wrapped in more detailed messages for
// debugging.
func CreateLockFile(lockFilePath string, verbose bool) (*flock.Flock, error) {
	fi, err := os.Stat(filepath.Dir(lockFilePath))
	if err != nil {
		return nil, errors.Wrap(err, "database directory is not available")
	}
	if !fi.IsDir() {
		return nil, errors.New("database directory is not a directory")
	}
	lock := flock.New(lockFilePath)
	ok, err := lock.TryLock()
	if err != nil {
		return nil, errors.Wrap(err, "error acquiring a lock")
	}
	if !ok {
		return nil, errors.Errorf("could not acquire lock on %s", lockFilePath)
	}
	if verbose {
		log.Printf("Acquired lock file lock (%s)", lockFilePath)
	}
	return lock, nil
}
