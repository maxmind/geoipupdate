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

const (
	extension     = ".mmdb"
	tempExtension = ".temporary"
)

// LocalFileWriter is a database.Writer that stores the database to the
// local file system.
type LocalFileWriter struct {
	dir              string
	preserveFileTime bool
	verbose          bool
}

// NewLocalFileWriter create a LocalFileWriter.
func NewLocalFileWriter(
	databaseDir string,
	preserveFileTime bool,
	verbose bool,
) (*LocalFileWriter, error) {
	err := os.MkdirAll(filepath.Dir(databaseDir), 0o750)
	if err != nil {
		return nil, fmt.Errorf("error creating database directory: %w", err)
	}

	return &LocalFileWriter{
		dir:              databaseDir,
		preserveFileTime: preserveFileTime,
		verbose:          verbose,
	}, nil
}

// Write writes the result struct returned by a Reader to a database file.
func (w *LocalFileWriter) Write(result *ReadResult) error {
	// exit early if we've got the latest database version.
	if strings.EqualFold(result.oldHash, result.newHash) {
		log.Printf("Database %s up to date", result.editionID)
		return nil
	}

	defer func() {
		if err := result.reader.Close(); err != nil {
			log.Printf("error closing reader for %s: %+v", result.editionID, err)
		}
	}()

	databaseFilePath := w.getFilePath(result.editionID)

	// write the Reader's result into a temporary file.
	fw, err := newFileWriter(databaseFilePath + tempExtension)
	if err != nil {
		return fmt.Errorf("error setting up database writer for %s: %w", result.editionID, err)
	}
	defer func() {
		if err := fw.close(); err != nil {
			log.Printf("error closing file writer: %+v", err)
		}
	}()

	if err := fw.write(result.reader); err != nil {
		return fmt.Errorf("error writing to the temp file for %s: %w", result.editionID, err)
	}

	// make sure the hash of the temp file matches the expected hash.
	if err := fw.validateHash(result.newHash); err != nil {
		return fmt.Errorf("error validating hash for %s: %w", result.editionID, err)
	}

	// move the temoporary database file into it's final location and
	// sync the directory.
	if err := fw.syncAndRename(databaseFilePath); err != nil {
		return fmt.Errorf("error renaming temp file: %w", err)
	}

	// sync database directory.
	if err := syncDir(filepath.Dir(databaseFilePath)); err != nil {
		return fmt.Errorf("error syncing database directory: %w", err)
	}

	// check if we need to set the file's modified at time
	if w.preserveFileTime {
		if err := setModifiedAtTime(databaseFilePath, result.modifiedAt); err != nil {
			return err
		}
	}

	if w.verbose {
		log.Printf("Database %s successfully updated: %+v", result.editionID, result.newHash)
	}

	return nil
}

// GetHash returns the hash of the current database file.
func (w *LocalFileWriter) GetHash(editionID string) (string, error) {
	databaseFilePath := w.getFilePath(editionID)
	//nolint:gosec // we really need to read this file.
	database, err := os.Open(databaseFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			if w.verbose {
				log.Print("Database does not exist, returning zeroed hash")
			}
			return ZeroMD5, nil
		}
		return "", fmt.Errorf("error opening database: %w", err)
	}

	defer func() {
		if err := database.Close(); err != nil {
			log.Println(fmt.Errorf("error closing database: %w", err))
		}
	}()

	md5Hash := md5.New()
	if _, err := io.Copy(md5Hash, database); err != nil {
		return "", fmt.Errorf("error calculating database hash: %w", err)
	}

	result := byteToString(md5Hash.Sum(nil))
	if w.verbose {
		log.Printf("Calculated MD5 sum for %s: %s", databaseFilePath, result)
	}
	return result, nil
}

// getFilePath construct the file path for a database edition.
func (w *LocalFileWriter) getFilePath(editionID string) string {
	return filepath.Join(w.dir, editionID) + extension
}

// fileWriter is used to write the content of a Reader's response
// into a file.
type fileWriter struct {
	// file is used for writing the Reader's response.
	file *os.File
	// md5Writer is used to verify the integrity of the received data.
	md5Writer hash.Hash
}

// newFileWriter initializes a new fileWriter struct.
func newFileWriter(path string) (*fileWriter, error) {
	// prepare temp file for initial writing.
	//nolint:gosec // we really need to read this file.
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return nil, fmt.Errorf("error creating temporary file at %s: %w", path, err)
	}

	return &fileWriter{
		file:      file,
		md5Writer: md5.New(),
	}, nil
}

// close closes and deletes the file.
func (w *fileWriter) close() error {
	if err := w.file.Close(); err != nil {
		var perr *os.PathError
		if !errors.As(err, &perr) || !errors.Is(perr.Err, os.ErrClosed) {
			return fmt.Errorf("error closing temporary file: %w", err)
		}
	}

	if err := os.Remove(w.file.Name()); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("error removing temporary file: %w", err)
	}

	return nil
}

// write writes the content of reader to the file.
func (w *fileWriter) write(r io.Reader) error {
	writer := io.MultiWriter(w.md5Writer, w.file)
	if _, err := io.Copy(writer, r); err != nil {
		return fmt.Errorf("error writing database: %w", err)
	}
	return nil
}

// validateHash validates the hash of the file against a known value.
func (w *fileWriter) validateHash(h string) error {
	tempFileHash := byteToString(w.md5Writer.Sum(nil))
	if !strings.EqualFold(h, tempFileHash) {
		return fmt.Errorf("md5 of new database (%s) does not match expected md5 (%s)", tempFileHash, h)
	}
	return nil
}

// syncAndRename syncs the content of the file to storage and renames it.
func (w *fileWriter) syncAndRename(name string) error {
	if err := w.file.Sync(); err != nil {
		return fmt.Errorf("error syncing temporary file: %w", err)
	}
	if err := w.file.Close(); err != nil {
		return fmt.Errorf("error closing temporary file: %w", err)
	}
	if err := os.Rename(w.file.Name(), name); err != nil {
		return fmt.Errorf("error moving database into place: %w", err)
	}
	return nil
}

// syncDir syncs the content of a directory to storage.
func syncDir(path string) error {
	// fsync the directory. http://austingroupbugs.net/view.php?id=672
	//nolint:gosec // we really need to read this file.
	d, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("error opening database directory %s: %w", path, err)
	}
	defer func() {
		if err := d.Close(); err != nil {
			log.Printf("closing directory %s: %+v", path, err)
		}
	}()

	// We ignore Sync errors as they primarily happen on file systems that do
	// not support sync.
	//nolint:errcheck // See above.
	_ = d.Sync()
	return nil
}

// setModifiedAtTime sets the times for a database file to a certain value.
func setModifiedAtTime(path string, t time.Time) error {
	if err := os.Chtimes(path, t, t); err != nil {
		return fmt.Errorf("error setting times on file %s: %w", path, err)
	}
	return nil
}

// byteToString returns the base16 representation of a byte array.
func byteToString(b []byte) string {
	return fmt.Sprintf("%x", b)
}
