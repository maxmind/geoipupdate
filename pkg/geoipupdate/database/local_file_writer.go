package database

import (
	"crypto/md5"
	"fmt"
	"github.com/gofrs/flock"
	"github.com/pkg/errors"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const zeroMD5 = "00000000000000000000000000000000"

type LocalFileDatabaseWriter struct {
	filePath   string
	lockFile   string
	verbose    bool
	lock       *flock.Flock
	oldHash    string
	targetFile *os.File
	swapFile   *os.File
}

func NewLocalFileDatabaseWriter(filePath string, lockFile string, verbose bool) (*LocalFileDatabaseWriter, error) {
	dbWriter := &LocalFileDatabaseWriter{
		filePath: filePath,
		lockFile: lockFile,
		verbose:  verbose,
	}
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			dbWriter.oldHash = zeroMD5
			dbWriter.targetFile, err = os.Create(filePath)
			if err != nil {
				return nil, errors.Wrap(err, "Encountered an error creating file "+filePath)
			}
		} else {
			return nil, errors.Wrap(err, "Received an unexpected error attempting to open file "+filePath)
		}
	} else {
		hash := md5.New()
		if _, err := io.Copy(hash, file); err != nil {
			return nil, errors.Wrap(err, "Encountered an error while createing hash for file "+filePath)
		}
		dbWriter.oldHash = fmt.Sprintf("%x", hash.Sum(nil))
		if verbose {
			log.Printf("Calculated MD5 sum for %s: %s", filePath, dbWriter.oldHash)
		}
	}
	if err := dbWriter.lockDirectory(); err != nil {
		return nil, err
	}

	dbWriter.swapFile, err = os.Create(fmt.Sprintf("%s.swap", dbWriter.filePath))
	if err != nil {
		return nil, errors.Wrap(err, "Error creating swap file")
	}

	return dbWriter, nil
}

func (writer *LocalFileDatabaseWriter) lockDirectory() error {
	fi, err := os.Stat(filepath.Dir(writer.filePath))
	if err != nil {
		return errors.Wrap(err, "database directory is not available")
	}
	if !fi.IsDir() {
		return errors.New("database directory is not a directory")
	}
	if writer.verbose {
		log.Printf("Acquired lock file lock (%s)", writer.lockFile)
	}
	writer.lock = flock.New(writer.lockFile)
	ok, err := writer.lock.TryLock()
	if err != nil {
		return errors.Wrap(err, "error acquiring a lock")
	}
	if !ok {
		return errors.Errorf("could not acquire lock on %s", writer.lockFile)
	}
	return nil
}

func (writer *LocalFileDatabaseWriter) Write(p []byte) (n int, err error) {
	return writer.swapFile.Write(p)
}

func (writer *LocalFileDatabaseWriter) Close() (err error) {
	return writer.swapFile.Close()
}

func (writer *LocalFileDatabaseWriter) ValidHash(expectedHash string) error {
	md5Writer := md5.New()
	reader, err := os.Open(writer.swapFile.Name())
	if err != nil {
		return errors.Wrap(err, "swap file was unable to be opened")
	}
	if _, err := io.Copy(md5Writer, reader); err != nil {
		return errors.Wrap(err, "no X-Database-MD5 header found")
	}
	hash := fmt.Sprintf("%x", md5Writer.Sum(nil))
	if !strings.EqualFold(hash, expectedHash) {
		return errors.Errorf("MD5 of new database (%s) does not match expected MD5 (%s)", hash, expectedHash)
	}
	return nil
}

func (writer *LocalFileDatabaseWriter) SetFileModificationTime(lastModified time.Time) error {
	if err := os.Chtimes(writer.filePath, lastModified, lastModified); err != nil {
		return errors.Wrap(err, "error setting times on file")
	}
	return nil
}

func (writer *LocalFileDatabaseWriter) Commit() error {
	if err := os.Rename(writer.swapFile.Name(), writer.filePath); err != nil {
		return errors.Wrap(err, "Error moving database into place")
	}

	_ = os.Remove(writer.swapFile.Name())

	// fsync the directory. http://austingroupbugs.net/view.php?id=672
	dh, err := os.Open(filepath.Dir(writer.filePath))
	if err != nil {
		return errors.Wrap(err, "error opening database directory")
	}
	defer func() {
		if err := dh.Close(); err != nil {
			log.Fatalf("Error closing directory: %+v", errors.Wrap(err, "closing directory"))
		}
	}()

	// We ignore Sync errors as they primarily happen on file systems that do
	// not support sync.
	_ = dh.Sync()
	_ = writer.lock.Unlock()
	return nil
}

func (writer *LocalFileDatabaseWriter) GetHash() (string, error) {
	return writer.oldHash, nil
}
