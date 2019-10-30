package database

import (
	"compress/gzip"
	"fmt"
	"github.com/maxmind/geoipupdate/pkg/geoipupdate"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

//HTTPDatabaseReader is a database.Reader that uses an HTTP client to retrieve the database data
type HTTPDatabaseReader struct {
	client            *http.Client
	url               string
	licenseKey        string
	accountID         int
	preserveFileTimes bool
	verbose           bool
}

func NewHTTPDatabaseReader(client *http.Client, config *geoipupdate.Config) Reader {
	return &HTTPDatabaseReader{
		client:            client,
		url:               config.URL,
		licenseKey:        config.LicenseKey,
		accountID:         config.AccountID,
		preserveFileTimes: config.PreserveFileTimes,
		verbose:           config.Verbose,
	}
}

//Get retrieves the data for a given editionID using an HTTP client to MaxMind, writes it to database.Writer,
// and validates the associated hash before committing
func (reader *HTTPDatabaseReader) Get(destination Writer, editionID string) error {
	defer func() {
		if err := destination.Close(); err != nil {
			log.Println(err)
		}
	}()
	lastHash, err := destination.GetHash()
	if err != nil {
		return errors.Wrap(err, "unable to get previous hash")
	}
	maxMindURL := fmt.Sprintf(
		"%s/geoip/databases/%s/update?db_md5=%s",
		reader.url,
		url.PathEscape(editionID),
		url.QueryEscape(lastHash),
	)

	req, err := http.NewRequest(http.MethodGet, maxMindURL, nil)
	if err != nil {
		return errors.Wrap(err, "error creating request")
	}
	if reader.accountID != 0 {
		req.SetBasicAuth(fmt.Sprintf("%d", reader.accountID), reader.licenseKey)
	}

	if reader.verbose {
		log.Printf("Performing update request to %s", maxMindURL)
	}
	response, err := reader.client.Do(req)
	if err != nil {
		return errors.Wrap(err, "error performing HTTP request")
	}
	defer func() {
		if err := response.Body.Close(); err != nil {
			log.Fatalf("Error closing response body: %+v", errors.Wrap(err, "closing body"))
		}
	}()

	if response.StatusCode == http.StatusNotModified {
		if reader.verbose {
			log.Printf("No new updates available for %s", editionID)
		}
		return nil
	}

	if response.StatusCode != http.StatusOK {
		buf, err := ioutil.ReadAll(io.LimitReader(response.Body, 256))
		if err == nil {
			return errors.Errorf("unexpected HTTP status code: %s: %s", response.Status, buf)
		}
		return errors.Errorf("unexpected HTTP status code: %s", response.Status)
	}

	newMD5 := response.Header.Get("X-Database-MD5")
	if newMD5 == "" {
		return errors.New("no X-Database-MD5 header found")
	}

	gzReader, err := gzip.NewReader(response.Body)
	if err != nil {
		return errors.Wrap(err, "encountered an error creating GZIP reader")
	}
	defer func() {
		if err := gzReader.Close(); err != nil {
			log.Printf("error closing gzip reader: %s", err)
		}
	}()

	if _, err = io.Copy(destination, gzReader); err != nil {
		return errors.Wrap(err, "encountered an error writing out MaxMind's response")
	}

	if err := destination.ValidHash(newMD5); err != nil {
		return err
	}

	if err := destination.Commit(); err != nil {
		return errors.Wrap(err, "encountered an issue committing database update")
	}

	if reader.preserveFileTimes {
		modificationTime, err := lastModified(response.Header.Get("Last-Modified"))
		if err != nil {
			return errors.Wrap(err, "unable to get last modified time")
		}
		err = destination.SetFileModificationTime(modificationTime)
		if err != nil {
			return errors.Wrap(err, "unable to set modification time")
		}
	}

	return nil
}

//LastModified retrieves the date that the MaxMind database was last modified
func lastModified(lastModified string) (time.Time, error) {
	if lastModified == "" {
		return time.Time{}, errors.New("no Last-Modified header found")
	}

	t, err := time.ParseInLocation(time.RFC1123, lastModified, time.UTC)
	if err != nil {
		return time.Time{}, errors.Wrap(err, "error parsing time")
	}

	return t, nil
}
