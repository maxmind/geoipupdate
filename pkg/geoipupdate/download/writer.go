package download

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/maxmind/geoipupdate/v6/pkg/geoipupdate/internal"
	"github.com/maxmind/geoipupdate/v6/pkg/geoipupdate/vars"
)

const (
	downloadEndpoint = "%s/geoip/databases/%s/download?date=%s&suffix=tar.gz"
)

// DownloadEdition downloads and writes an edition to a database file.
func (d *Download) DownloadEdition(ctx context.Context, edition Metadata) error {
	date := strings.ReplaceAll(edition.Date, "-", "")
	requestURL := fmt.Sprintf(downloadEndpoint, d.url, edition.EditionID, date)
	if d.verbose {
		log.Printf("Downloading: %s", requestURL)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Add("User-Agent", "geoipupdate/"+vars.Version)
	req.SetBasicAuth(strconv.Itoa(d.accountID), d.licenseKey)

	response, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("performing HTTP request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		responseBody, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return fmt.Errorf("reading response body: %w", err)
		}

		errResponse := internal.ResponseError{
			StatusCode: response.StatusCode,
		}

		if err := json.Unmarshal(responseBody, &errResponse); err != nil {
			errResponse.Message = err.Error()
		}

		return fmt.Errorf("requesting edition: %w", errResponse)
	}

	databaseFilePath := d.getFilePath(edition.EditionID)

	// write the result into a temporary file.
	fw, err := newFileWriter(databaseFilePath + ".temporary")
	if err != nil {
		return fmt.Errorf("setting up database writer for %s: %w", edition.EditionID, err)
	}
	defer func() {
		if closeErr := fw.close(); closeErr != nil {
			err = errors.Join(
				err,
				fmt.Errorf("closing file writer: %w", closeErr),
			)
		}
	}()

	gzReader, err := gzip.NewReader(response.Body)
	if err != nil {
		return fmt.Errorf("encountered an error creating GZIP reader: %w", err)
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	// iterate through the tar archive to extract the mmdb file
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			return errors.New("tar archive does not contain an mmdb file")
		}
		if err != nil {
			return fmt.Errorf("reading tar archive: %w", err)
		}

		if strings.HasSuffix(header.Name, Extension) {
			break
		}
	}

	if err = fw.write(tarReader); err != nil {
		return fmt.Errorf("writing to the temp file for %s: %w", edition.EditionID, err)
	}

	// make sure the hash of the temp file matches the expected hash.
	if err = fw.validateHash(edition.MD5); err != nil {
		return fmt.Errorf("validating hash for %s: %w", edition.EditionID, err)
	}

	// move the temoporary database file into it's final location and
	// sync the directory.
	if err = fw.syncAndRename(databaseFilePath); err != nil {
		return fmt.Errorf("renaming temp file: %w", err)
	}

	// sync database directory.
	if err = syncDir(filepath.Dir(databaseFilePath)); err != nil {
		return fmt.Errorf("syncing database directory: %w", err)
	}

	// check if we need to set the file's modified at time
	if d.preserveFileTimes {
		if err = setModifiedAtTime(databaseFilePath, edition.Date); err != nil {
			return err
		}
	}

	if d.verbose {
		log.Printf("Database %s successfully updated: %+v", edition.EditionID, edition.MD5)
	}

	return nil
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
		return nil, fmt.Errorf("creating temporary file at %s: %w", path, err)
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
			return fmt.Errorf("closing temporary file: %w", err)
		}
	}

	err := os.Remove(w.file.Name())
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("removing temporary file: %w", err)
	}

	return nil
}

// write writes the content of reader to the file.
func (w *fileWriter) write(r io.Reader) error {
	writer := io.MultiWriter(w.md5Writer, w.file)
	if _, err := io.Copy(writer, r); err != nil {
		return fmt.Errorf("writing database: %w", err)
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
		return fmt.Errorf("syncing temporary file: %w", err)
	}
	if err := w.file.Close(); err != nil {
		return fmt.Errorf("closing temporary file: %w", err)
	}
	if err := os.Rename(w.file.Name(), name); err != nil {
		return fmt.Errorf("moving database into place: %w", err)
	}
	return nil
}

// syncDir syncs the content of a directory to storage.
func syncDir(path string) error {
	// fsync the directory. https://austingroupbugs.net/view.php?id=672
	//nolint:gosec // we really need to read this file.
	d, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("opening database directory %s: %w", path, err)
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
func setModifiedAtTime(path string, dateString string) error {
	releaseDate, err := ParseTime(dateString)
	if err != nil {
		return err
	}

	if err := os.Chtimes(path, releaseDate, releaseDate); err != nil {
		return fmt.Errorf("setting times on file %s: %w", path, err)
	}
	return nil
}
