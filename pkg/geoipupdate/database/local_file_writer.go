package database

import (
	"crypto/md5"
	"fmt"
	"github.com/gofrs/flock"
	"github.com/pkg/errors"
	"hash"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const zeroMD5 = "00000000000000000000000000000000"

//LocalFileDatabaseWriter is a database.Writer that stores the database to the local file system
type LocalFileDatabaseWriter struct {
	filePath      string
	lockFile      string
	verbose       bool
	lock          *flock.Flock
	oldHash       string
	fileWriter    io.Writer
	temporaryFile *os.File
	md5Writer     hash.Hash
}

//NewLocalFileDatabaseWriter create a new LocalFileDatabaseWriter, creating necessary lock and temporary files to protect
// the database from concurrent writes
func NewLocalFileDatabaseWriter(filePath string, lockFile string, verbose bool) (*LocalFileDatabaseWriter, error) {
	dbWriter := &LocalFileDatabaseWriter{
		filePath: filePath,
		lockFile: lockFile,
		verbose:  verbose,
	}
	if err := dbWriter.createOldMD5Hash(); err != nil {
		return nil, err
	}
	if err := dbWriter.createLockFile(); err != nil {
		return nil, err
	}

	var err error
	temporaryFilename := fmt.Sprintf("%s.temporary", dbWriter.filePath)
	dbWriter.temporaryFile, err = os.OpenFile(temporaryFilename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return nil, errors.Wrap(err, "error creating temporary file")
	}
	dbWriter.md5Writer = md5.New()
	dbWriter.fileWriter = io.MultiWriter(dbWriter.md5Writer, dbWriter.temporaryFile)

	return dbWriter, nil
}

func (writer *LocalFileDatabaseWriter) createOldMD5Hash() error {
	currentDatabaseFile, err := os.Open(writer.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			writer.oldHash = zeroMD5
			return nil
		}
		return errors.Wrap(err, "received an unexpected error attempting to open temporaryFile "+writer.filePath)
	}

	defer func() {
		err := currentDatabaseFile.Close()
		if err != nil {
			log.Println(errors.Wrap(err, "error closing current database temporaryFile "+writer.filePath))
		}
	}()
	oldHash := md5.New()
	if _, err := io.Copy(oldHash, currentDatabaseFile); err != nil {
		return errors.Wrap(err, "encountered an error while creating oldHash for temporaryFile "+writer.filePath)
	}
	writer.oldHash = fmt.Sprintf("%x", oldHash.Sum(nil))
	if writer.verbose {
		log.Printf("Calculated MD5 sum for %s: %s", writer.filePath, writer.oldHash)
	}
	return nil
}

func (writer *LocalFileDatabaseWriter) createLockFile() error {
	fi, err := os.Stat(filepath.Dir(writer.filePath))
	if err != nil {
		return errors.Wrap(err, "database directory is not available")
	}
	if !fi.IsDir() {
		return errors.New("database directory is not a directory")
	}
	writer.lock = flock.New(writer.lockFile)
	ok, err := writer.lock.TryLock()
	if err != nil {
		return errors.Wrap(err, "error acquiring a lock")
	}
	if !ok {
		return errors.Errorf("could not acquire lock on %s", writer.lockFile)
	}
	if writer.verbose {
		log.Printf("Acquired lock file lock (%s)", writer.lockFile)
	}
	return nil
}

//Write writes data to temporary file
func (writer *LocalFileDatabaseWriter) Write(p []byte) (int, error) {
	return writer.fileWriter.Write(p)
}

//Close closes the temporary file
func (writer *LocalFileDatabaseWriter) Close() error {
	if err := writer.temporaryFile.Close(); err != nil && errors.Cause(err) == os.ErrClosed {
		return errors.Wrap(err, "error closing temporary file")
	}
	if err := os.Remove(writer.temporaryFile.Name()); err != nil && errors.Cause(err) == os.ErrNotExist {
		return errors.Wrap(err, "error removing temporary file")
	}
	if err := writer.lock.Unlock(); err != nil {
		return errors.Wrap(err, "error releasing lock file")
	}
	return nil
}

//ValidHash checks that the temporary file's MD5 matches the expectedHash
func (writer *LocalFileDatabaseWriter) ValidHash(expectedHash string) error {
	actualHash := fmt.Sprintf("%x", writer.md5Writer.Sum(nil))
	if !strings.EqualFold(actualHash, expectedHash) {
		return errors.Errorf("md5 of new database (%s) does not match expected md5 (%s)", actualHash, expectedHash)
	}
	return nil
}

//SetFileModificationTime explicitly sets the database's file write time to the provided time
func (writer *LocalFileDatabaseWriter) SetFileModificationTime(lastModified time.Time) error {
	if err := os.Chtimes(writer.filePath, lastModified, lastModified); err != nil {
		return errors.Wrap(err, "error setting times on file")
	}
	return nil
}

//Commit renames the temporary file to the name of the database file before syncing the directory
func (writer *LocalFileDatabaseWriter) Commit() error {
	if err := writer.temporaryFile.Sync(); err != nil {
		return errors.Wrap(err, "error syncing temporary file")
	}
	if err := writer.temporaryFile.Close(); err != nil {
		return errors.Wrap(err, "error closing temporary file")
	}
	if err := os.Rename(writer.temporaryFile.Name(), writer.filePath); err != nil {
		return errors.Wrap(err, "error moving database into place")
	}

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
	return nil
}

//GetHash returns the hash of the current database file
func (writer *LocalFileDatabaseWriter) GetHash() (string, error) {
	return writer.oldHash, nil
}
