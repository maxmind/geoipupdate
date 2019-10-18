package geoipupdate

import (
	"bytes"
	"fmt"
	"github.com/maxmind/geoipupdate/pkg/geoipupdate/database"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
)

//Run takes the information from a Config and copies all of the provided EditionIDs from a database.Reader to a database.Writer
func Run(
	config *Config,
) error {
	client := buildClient(config)
	dbReader := &database.HTTPDatabaseReader{
		Client:     client,
		URL:        config.URL,
		LicenseKey: config.LicenseKey,
		AccountID:  config.AccountID,
		Verbose:    config.Verbose,
	}
	for _, editionID := range config.EditionIDs {
		filename, err := getFileName(config, editionID, client)
		if err != nil {
			return errors.Wrap(err, "error retrieving filename")
		}
		filePath := filepath.Join(config.DatabaseDirectory, filename)
		dbWriter, err := database.NewLocalFileDatabaseWriter(filePath, config.LockFile, config.Verbose)
		if err != nil {
			return errors.Wrap(err, "Error create database writer")
		}
		if err := UpdateEdition(dbReader, dbWriter, config, editionID); err != nil {
			return err
		}
	}
	return nil
}

func buildClient(
	config *Config,
) *http.Client {
	var client *http.Client
	if config.Proxy != nil {
		client = &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(config.Proxy)}}
	} else {
		client = &http.Client{}
	}
	return client
}

func getFileName(
	config *Config,
	editionID string,
	client *http.Client,
) (string, error) {
	url := fmt.Sprintf(
		"%s/app/update_getfilename?product_id=%s",
		config.URL,
		url.QueryEscape(editionID),
	)

	if config.Verbose {
		log.Printf("Performing get filename request to %s", url)
	}
	res, err := client.Get(url)
	if err != nil {
		return "", errors.Wrap(err, "error performing HTTP request")
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			log.Fatalf("Error closing response body: %+v", errors.Wrap(err, "closing body"))
		}
	}()

	buf, err := ioutil.ReadAll(io.LimitReader(res.Body, 256))
	if err != nil {
		return "", errors.Wrap(err, "error reading response body")
	}

	if res.StatusCode != http.StatusOK {
		return "", errors.Errorf("unexpected HTTP status code: %s: %s", res.Status, buf)
	}

	if len(buf) == 0 {
		return "", errors.New("response body is empty")
	}

	if bytes.Count(buf, []byte("\n")) > 0 ||
		bytes.Count(buf, []byte("\x00")) > 0 {
		return "", errors.New("invalid characters in filename")
	}

	return string(buf), nil
}

//UpdateEdition copies the contents of a database.Reader to a single database.Writer
func UpdateEdition(dbReader database.Reader, dbWriter database.Writer, config *Config, editionID string) error {
	if err := dbReader.Get(dbWriter, editionID); err != nil {
		return errors.Wrap(err, "error updating")
	}
	if config.PreserveFileTimes {
		modificationTime, err := dbReader.LastModified()
		if err != nil {
			return errors.Wrap(err, "Unable to get last modified time")
		}
		err = dbWriter.SetFileModificationTime(modificationTime)
		if err != nil {
			return errors.Wrap(err, "Unable to set modification time")
		}
	}
	return nil
}
