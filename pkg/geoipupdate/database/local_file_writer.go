package database

import (
	"crypto/md5"
	"errors"
	"fmt"
	"hash"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// LocalFileDatabaseWriter is a database.Writer that stores the database to the
// local file system.
type LocalFileDatabaseWriter struct {
	filePath      string
	verbose       bool
	oldHash       string
	fileWriter    io.Writer
	temporaryFile *os.File
	md5Writer     hash.Hash
}

// NewLocalFileDatabaseWriter create a LocalFileDatabaseWriter. It creates the
// necessary lock and temporary files to protect the database from concurrent
// writes.
func NewLocalFileDatabaseWriter(filePath string, lock *FileLock, verbose bool) (*LocalFileDatabaseWriter, error) {
	dbWriter := &LocalFileDatabaseWriter{
		filePath: filePath,
		verbose:  verbose,
	}

	if _, err := lock.acquireLock(); err != nil {
		return nil, err
	}

	if err := dbWriter.createOldMD5Hash(); err != nil {
		return nil, err
	}

	var err error
	temporaryFilename := fmt.Sprintf("%s.temporary", dbWriter.filePath)
	//nolint:gosec // We want the permission to be world readable
	dbWriter.temporaryFile, err = os.OpenFile(
		temporaryFilename,
		os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
		0o644,
	)
	if err != nil {
		return nil, fmt.Errorf("error creating temporary file: %w", err)
	}
	dbWriter.md5Writer = md5.New()
	dbWriter.fileWriter = io.MultiWriter(dbWriter.md5Writer, dbWriter.temporaryFile)

	return dbWriter, nil
}

func (writer *LocalFileDatabaseWriter) createOldMD5Hash() error {
	currentDatabaseFile, err := os.Open(writer.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			writer.oldHash = ZeroMD5
			return nil
		}
		return fmt.Errorf("error opening database: %w", err)
	}

	defer func() {
		err := currentDatabaseFile.Close()
		if err != nil {
			log.Println(fmt.Errorf("error closing database: %w", err))
		}
	}()
	oldHash := md5.New()
	if _, err := io.Copy(oldHash, currentDatabaseFile); err != nil {
		return fmt.Errorf("error calculating database hash: %w", err)
	}
	writer.oldHash = fmt.Sprintf("%x", oldHash.Sum(nil))
	if writer.verbose {
		log.Printf("Calculated MD5 sum for %s: %s", writer.filePath, writer.oldHash)
	}
	return nil
}

// Write writes to the temporary file.
func (writer *LocalFileDatabaseWriter) Write(p []byte) (int, error) {
	n, err := writer.fileWriter.Write(p)
	if err != nil {
		return 0, fmt.Errorf("error writing: %w", err)
	}
	return n, nil
}

// Close closes the temporary file.
func (writer *LocalFileDatabaseWriter) Close() error {
	err := writer.temporaryFile.Close()
	if err != nil {
		var perr *os.PathError
		if !errors.As(err, &perr) || !errors.Is(perr.Err, os.ErrClosed) {
			return fmt.Errorf("error closing temporary file: %w", err)
		}
	}

	if err := os.Remove(writer.temporaryFile.Name()); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("error removing temporary file: %w", err)
	}
	return nil
}

// ValidHash checks that the temporary file's MD5 matches the given hash.
func (writer *LocalFileDatabaseWriter) ValidHash(expectedHash string) error {
	actualHash := fmt.Sprintf("%x", writer.md5Writer.Sum(nil))
	if !strings.EqualFold(actualHash, expectedHash) {
		return fmt.Errorf("md5 of new database (%s) does not match expected md5 (%s)", actualHash, expectedHash)
	}
	return nil
}

// SetFileModificationTime sets the database's file access and modified times
// to the given time.
func (writer *LocalFileDatabaseWriter) SetFileModificationTime(lastModified time.Time) error {
	if err := os.Chtimes(writer.filePath, lastModified, lastModified); err != nil {
		return fmt.Errorf("error setting times on file: %w", err)
	}
	return nil
}

// Commit renames the temporary file to the name of the database file and syncs
// the directory.
func (writer *LocalFileDatabaseWriter) Commit() error {
	if err := writer.temporaryFile.Sync(); err != nil {
		return fmt.Errorf("error syncing temporary file: %w", err)
	}
	if err := writer.temporaryFile.Close(); err != nil {
		return fmt.Errorf("error closing temporary file: %w", err)
	}
	if err := os.Rename(writer.temporaryFile.Name(), writer.filePath); err != nil {
		return fmt.Errorf("error moving database into place: %w", err)
	}

	// fsync the directory. http://austingroupbugs.net/view.php?id=672
	dh, err := os.Open(filepath.Dir(writer.filePath))
	if err != nil {
		return fmt.Errorf("error opening database directory: %w", err)
	}

	// We ignore Sync errors as they primarily happen on file systems that do
	// not support sync.
	//nolint:errcheck // See above.
	_ = dh.Sync()

	if err := dh.Close(); err != nil {
		return fmt.Errorf("closing directory: %w", err)
	}
	return nil
}

// GetHash returns the hash of the current database file.
func (writer *LocalFileDatabaseWriter) GetHash() string {
	return writer.oldHash
}
