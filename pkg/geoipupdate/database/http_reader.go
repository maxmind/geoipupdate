// Package database provides an abstraction over getting and writing a
// database file.
package database

import (
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/maxmind/geoipupdate/v4/pkg/geoipupdate"
	"github.com/maxmind/geoipupdate/v4/pkg/geoipupdate/internal"
	"golang.org/x/sync/errgroup"
)

// HTTPDatabaseReader is a Reader that uses an HTTP client to retrieve
// databases.
type HTTPDatabaseReader struct {
	client            *http.Client
	retryFor          time.Duration
	url               string
	licenseKey        string
	accountID         int
	preserveFileTimes bool
	verbose           bool

	// processor is used to parallelize and limit the number of workers
	// processing read requests.
	processor *errgroup.Group
	// processorFunc is the function that each worker runs to process requests.
	// It will mostly be customized for test purposes.
	processorFunc func(Writer, string) error
	// ctx is the context associated with the lifecycle of the download queue.
	//nolint: containedctx // context is stored here to keep the api intact.
	//                        it should be removed and properly propagated
	//                        in a future refactor.
	ctx context.Context
	// cancel cancels the context and stops the processing of the queue.
	cancel context.CancelFunc
}

// NewHTTPDatabaseReader creates a Reader that downloads database updates via
// HTTP.
func NewHTTPDatabaseReader(client *http.Client, config *geoipupdate.Config) Reader {
	ctx, cancel := context.WithCancel(context.Background())
	processor := new(errgroup.Group)
	processor.SetLimit(config.Parallelism)

	h := &HTTPDatabaseReader{
		client:            client,
		retryFor:          config.RetryFor,
		url:               config.URL,
		licenseKey:        config.LicenseKey,
		accountID:         config.AccountID,
		preserveFileTimes: config.PreserveFileTimes,
		verbose:           config.Verbose,
		processor:         processor,
		ctx:               ctx,
		cancel:            cancel,
	}
	h.processorFunc = func(w Writer, s string) error {
		return h.Get(w, s)
	}
	return h
}

// Queue is a wrapper around `Get` that starts the read job as a background
// process. It makes use of a errgroup.Group to limit the number of
// concurrent requests handled at a certain time.
func (reader *HTTPDatabaseReader) Queue(destination Writer, editionID string) {
	reader.processor.Go(func() error {
		return reader.processorFunc(destination, editionID)
	})
}

// Wait waits for the queue to finish processing and returns the first
// error encountered, if any.
func (reader *HTTPDatabaseReader) Wait() error {
	if err := reader.processor.Wait(); err != nil {
		return fmt.Errorf("error downloading database: %w", err)
	}
	return nil
}

// Stop cancels the execution of the queue.
func (reader *HTTPDatabaseReader) Stop() error {
	reader.cancel()
	return reader.Wait()
}

// Get retrieves the given edition ID using an HTTP client, writes it to the
// Writer, and validates the hash before committing.
func (reader *HTTPDatabaseReader) Get(destination Writer, editionID string) error {
	if err := reader.ctx.Err(); err != nil {
		return fmt.Errorf("error processing %s: %w", editionID, err)
	}

	defer func() {
		if err := destination.Close(); err != nil {
			log.Println(err)
		}
	}()

	updateURL := fmt.Sprintf(
		"%s/geoip/databases/%s/update?db_md5=%s",
		reader.url,
		url.PathEscape(editionID),
		url.QueryEscape(destination.GetHash()),
	)

	var modified bool
	// It'd be nice to not use a temporary file here. However the Writer API does
	// not currently support multiple attempts to download the file (it assumes
	// we'll begin writing once). Using a temporary file here should be a less
	// disruptive alternative to changing the API. If we change that API in the
	// future, adding something like Reset() may be desirable.
	tempFile, err := ioutil.TempFile("", "geoipupdate")
	if err != nil {
		return fmt.Errorf("error opening temporary file: %w", err)
	}
	defer func() {
		if err := tempFile.Close(); err != nil {
			log.Printf("error closing temporary file: %s", err)
		}
		if err := os.Remove(tempFile.Name()); err != nil {
			log.Printf("error removing temporary file: %s", err)
		}
	}()
	var newMD5 string
	var modificationTime time.Time
	err = internal.RetryWithBackoff(
		func() error {
			if reader.verbose {
				log.Printf("Performing update request to %s", updateURL)
			}

			newMD5, modificationTime, modified, err = reader.download(
				updateURL,
				editionID,
				tempFile,
			)
			return err
		},
		reader.retryFor,
	)
	if err != nil {
		return err
	}

	if !modified {
		return nil
	}

	if _, err := tempFile.Seek(0, 0); err != nil {
		return fmt.Errorf("error seeking: %w", err)
	}

	if _, err = io.Copy(destination, tempFile); err != nil {
		return fmt.Errorf("error writing response: %w", err)
	}

	if err := destination.ValidHash(newMD5); err != nil {
		return err
	}

	if err := destination.Commit(); err != nil {
		return fmt.Errorf("encountered an issue committing database update: %w", err)
	}

	if reader.preserveFileTimes {
		err = destination.SetFileModificationTime(modificationTime)
		if err != nil {
			return fmt.Errorf("unable to set modification time: %w", err)
		}
	}

	return nil
}

func (reader *HTTPDatabaseReader) download(
	updateURL,
	editionID string,
	tempFile *os.File,
) (string, time.Time, bool, error) {
	// Prepare a clean slate for this download attempt.

	if err := tempFile.Truncate(0); err != nil {
		return "", time.Time{}, false, fmt.Errorf("error truncating: %w", err)
	}
	if _, err := tempFile.Seek(0, 0); err != nil {
		return "", time.Time{}, false, fmt.Errorf("error seeking: %w", err)
	}

	// Perform the download.
	//nolint: noctx // using the context would require an API change
	req, err := http.NewRequest(http.MethodGet, updateURL, nil)
	if err != nil {
		return "", time.Time{}, false, fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Add("User-Agent", "geoipupdate/"+geoipupdate.Version)
	req.SetBasicAuth(fmt.Sprintf("%d", reader.accountID), reader.licenseKey)

	response, err := reader.client.Do(req)
	if err != nil {
		return "", time.Time{}, false, fmt.Errorf("error performing HTTP request: %w", err)
	}

	defer response.Body.Close()

	if response.StatusCode == http.StatusNotModified {
		if reader.verbose {
			log.Printf("No new updates available for %s", editionID)
		}
		return "", time.Time{}, false, nil
	}

	if response.StatusCode != http.StatusOK {
		//nolint:errcheck // we are already returning an error.
		buf, _ := ioutil.ReadAll(io.LimitReader(response.Body, 256))
		httpErr := internal.HTTPError{
			Body:       string(buf),
			StatusCode: response.StatusCode,
		}
		return "", time.Time{}, false, fmt.Errorf("unexpected HTTP status code: %w", httpErr)
	}

	gzReader, err := gzip.NewReader(response.Body)
	if err != nil {
		return "", time.Time{}, false, fmt.Errorf("encountered an error creating GZIP reader: %w", err)
	}
	defer func() {
		if err := gzReader.Close(); err != nil {
			log.Printf("error closing gzip reader: %s", err)
		}
	}()

	//nolint:gosec // A decompression bomb is unlikely here
	if _, err := io.Copy(tempFile, gzReader); err != nil {
		return "", time.Time{}, false, fmt.Errorf("error writing response: %w", err)
	}

	newMD5 := response.Header.Get("X-Database-MD5")
	if newMD5 == "" {
		return "", time.Time{}, false, errors.New("no X-Database-MD5 header found")
	}

	modificationTime, err := lastModified(response.Header.Get("Last-Modified"))
	if err != nil {
		return "", time.Time{}, false, fmt.Errorf("unable to get last modified time: %w", err)
	}

	return newMD5, modificationTime, true, nil
}

// LastModified retrieves the date that the MaxMind database was last modified.
func lastModified(lastModified string) (time.Time, error) {
	if lastModified == "" {
		return time.Time{}, errors.New("no Last-Modified header found")
	}

	t, err := time.ParseInLocation(time.RFC1123, lastModified, time.UTC)
	if err != nil {
		return time.Time{}, fmt.Errorf("error parsing time: %w", err)
	}

	return t, nil
}
