// Package geoipupdate provides a library for using MaxMind's GeoIP Update
// service.
package geoipupdate

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/cenkalti/backoff/v4"

	"github.com/maxmind/geoipupdate/v6/pkg/geoipupdate/download"
	"github.com/maxmind/geoipupdate/v6/pkg/geoipupdate/internal"
)

// Client uses config data to initiate a download or update
// process for GeoIP databases.
type Client struct {
	config     *Config
	downloader download.Downloader
	output     *log.Logger
}

// NewClient initialized a new Client struct.
func NewClient(config *Config) (*Client, error) {
	d, err := download.New(
		config.AccountID,
		config.LicenseKey,
		config.URL,
		config.Proxy,
		config.DatabaseDirectory,
		config.PreserveFileTimes,
		config.EditionIDs,
		config.Verbose,
	)
	if err != nil {
		return nil, fmt.Errorf("initializing database downloader: %w", err)
	}

	return &Client{
		config:     config,
		downloader: d,
		output:     log.New(os.Stdout, "", 0),
	}, nil
}

// Run starts the download or update process.
func (c *Client) Run(ctx context.Context) error {
	fileLock, err := internal.NewFileLock(c.config.LockFile, c.config.Verbose)
	if err != nil {
		return fmt.Errorf("initializing file lock: %w", err)
	}
	if err := fileLock.Acquire(); err != nil {
		return fmt.Errorf("acquiring file lock: %w", err)
	}
	defer func() {
		if err := fileLock.Release(); err != nil {
			log.Printf("releasing file lock: %s", err)
		}
	}()

	var outdatedEditions []download.Metadata
	getOutdatedEditions := func() (err error) {
		outdatedEditions, err = c.downloader.GetOutdatedEditions(ctx)
		if err != nil {
			return fmt.Errorf("getting outdated database editions: %w", err)
		}
		return nil
	}

	if err := c.retry(getOutdatedEditions, "Couldn't get download metadata"); err != nil {
		return fmt.Errorf("getting download metadata: %w", err)
	}

	downloadEdition := func(edition download.Metadata) error {
		if err := c.downloader.DownloadEdition(ctx, edition); err != nil {
			return fmt.Errorf("downloading edition '%s': %w", edition.EditionID, err)
		}
		return nil
	}

	jobProcessor := internal.NewJobProcessor(ctx, c.config.Parallelism)
	for _, edition := range outdatedEditions {
		edition := edition
		processFunc := func(_ context.Context) error {
			err := c.retry(
				func() error { return downloadEdition(edition) },
				"Couldn't download "+edition.EditionID,
			)
			if err != nil {
				return err
			}
			return nil
		}
		jobProcessor.Add(processFunc)
	}

	// Run blocks until all jobs are processed or exits early after
	// the first encountered error.
	if err := jobProcessor.Run(ctx); err != nil {
		return fmt.Errorf("running the job processor: %w", err)
	}

	if c.config.Output {
		result, err := c.downloader.MakeOutput()
		if err != nil {
			return fmt.Errorf("marshaling result log: %w", err)
		}
		c.output.Print(string(result))
	}

	return nil
}

// retry implements a retry functionality for downloads for non permanent errors.
func (c *Client) retry(
	f func() error,
	logMsg string,
) error {
	// RetryFor value of 0 means that no retries should be performed.
	// Max zero retries has to be set to achieve that
	// because the backoff never stops if MaxElapsedTime is zero.
	exp := backoff.NewExponentialBackOff()
	exp.MaxElapsedTime = c.config.RetryFor
	b := backoff.BackOff(exp)
	if exp.MaxElapsedTime == 0 {
		b = backoff.WithMaxRetries(exp, 0)
	}

	return backoff.RetryNotify(
		func() error {
			if err := f(); err != nil {
				if internal.IsPermanentError(err) {
					return backoff.Permanent(err)
				}
				return err
			}
			return nil
		},
		b,
		func(err error, d time.Duration) {
			if c.config.Verbose {
				log.Printf("%s, retrying in %v: %v", logMsg, d, err)
			}
		},
	)
}
