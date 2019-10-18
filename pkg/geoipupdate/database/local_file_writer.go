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

//LocalFileDatabaseWriter is a database.Writer that stores the database to the local file system
type LocalFileDatabaseWriter struct {
	filePath string
	lockFile string
	verbose  bool
	lock     *flock.Flock
	oldHash  string
	swapFile *os.File
}

//NewLocalFileDatabaseWriter create a new LocalFileDatabaseWriter, creating necessary lock and swap files to protect
// the database from concurrent writes
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
		} else {
			return nil, errors.Wrap(err, "Received an unexpected error attempting to open file "+filePath)
		}
	} else {
		defer func() {
			err := file.Close()
			if err != nil {
				log.Println(errors.Wrap(err, "Error closing current datbase file "+filePath))
			}
		}()
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

//Write writes data to swap file
func (writer *LocalFileDatabaseWriter) Write(p []byte) (n int, err error) {
	return writer.swapFile.Write(p)
}

//Close closes the swap file
func (writer *LocalFileDatabaseWriter) Close() (err error) {
	_ = writer.swapFile.Close()
	_ = os.Remove(writer.swapFile.Name())
	if err := writer.lock.Unlock(); err != nil {
		return errors.Wrap(err, "error releasing lock file")
	}
	return nil
}

//ValidHash checks that the swap file's MD5 matches the expectedHash
func (writer *LocalFileDatabaseWriter) ValidHash(expectedHash string) error {
	_, _ = writer.swapFile.Seek(0, 0)
	md5Writer := md5.New()
	reader, err := os.Open(writer.swapFile.Name())
	defer func() {
		if err := reader.Close(); err != nil {
			log.Println(errors.Wrap(err, "Encountered an error closing swap file after validating its hash"))
		}
	}()
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

//SetFileModificationTime explicitly sets the database's file write time to the provided time
func (writer *LocalFileDatabaseWriter) SetFileModificationTime(lastModified time.Time) error {
	if err := os.Chtimes(writer.filePath, lastModified, lastModified); err != nil {
		return errors.Wrap(err, "error setting times on file")
	}
	return nil
}

//Commit renames the swap file to the name of the database file before syncing the directory
func (writer *LocalFileDatabaseWriter) Commit() error {
	_ = writer.swapFile.Close()
	if err := os.Rename(writer.swapFile.Name(), writer.filePath); err != nil {
		return errors.Wrap(err, "Error moving database into place")
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
