// Package geoipupdate provides a library for using MaxMind's GeoIP Update
// service.
package geoipupdate

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/cenkalti/backoff/v4"

	"github.com/maxmind/geoipupdate/v6/pkg/geoipupdate/api"
	"github.com/maxmind/geoipupdate/v6/pkg/geoipupdate/config"
	"github.com/maxmind/geoipupdate/v6/pkg/geoipupdate/internal"
	"github.com/maxmind/geoipupdate/v6/pkg/geoipupdate/lock"
	geoipupdatewriter "github.com/maxmind/geoipupdate/v6/pkg/geoipupdate/writer"
)

// Client uses config data to initiate a download or update
// process for GeoIP databases.
type Client struct {
	editionIDs  []string
	parallelism int
	retryFor    time.Duration
	printOutput bool

	downloader api.DownloadAPI
	locker     lock.Lock
	output     *log.Logger
	writer     geoipupdatewriter.Writer
}

// NewClient initialized a new Client struct.
func NewClient(
	conf *config.Config,
	downloader api.DownloadAPI,
	locker lock.Lock,
	writer geoipupdatewriter.Writer,
) *Client {
	return &Client{
		editionIDs:  conf.EditionIDs,
		parallelism: conf.Parallelism,
		printOutput: conf.Output,
		retryFor:    conf.RetryFor,

		downloader: downloader,
		locker:     locker,
		output:     log.New(os.Stdout, "", 0),
		writer:     writer,
	}
}

// Run starts the download or update process.
func (c *Client) Run(ctx context.Context) error {
	if err := c.locker.Acquire(); err != nil {
		return fmt.Errorf("acquiring file lock: %w", err)
	}
	slog.Debug("file lock acquired")
	defer func() {
		if err := c.locker.Release(); err != nil {
			log.Printf("releasing file lock: %s", err)
		}
		slog.Debug("file lock successfully released")
	}()

	oldEditionsHash := map[string]string{}
	for _, e := range c.editionIDs {
		hash, err := c.writer.GetHash(e)
		if err != nil {
			return fmt.Errorf("getting existing %q database hash: %w", e, err)
		}
		oldEditionsHash[e] = hash
		slog.Debug("existing database md5", "edition", e, "md5", hash)
	}

	var allEditions []api.Metadata
	getMetadata := func() (err error) {
		slog.Debug("requesting metadata")
		allEditions, err = c.downloader.GetMetadata(ctx, c.editionIDs)
		if err != nil {
			return fmt.Errorf("getting outdated database editions: %w", err)
		}
		return nil
	}

	if err := c.retry(getMetadata, "Couldn't get download metadata"); err != nil {
		return fmt.Errorf("getting download metadata: %w", err)
	}

	var outdatedEditions []api.Metadata
	for _, m := range allEditions {
		if m.MD5 != oldEditionsHash[m.EditionID] {
			outdatedEditions = append(outdatedEditions, m)
			continue
		}
		slog.Debug("database up to date", "edition", m.EditionID)
	}

	downloadEdition := func(edition api.Metadata) error {
		slog.Debug("downloading", "edition", edition.EditionID)
		reader, cleanupCallback, err := c.downloader.GetEdition(ctx, edition)
		if err != nil {
			return fmt.Errorf("downloading edition '%s': %w", edition.EditionID, err)
		}
		slog.Debug("writing", "edition", edition.EditionID)
		if err := c.writer.Write(edition, reader); err != nil {
			return fmt.Errorf("writing edition '%s': %w", edition.EditionID, err)
		}
		cleanupCallback()
		slog.Debug("database successfully downloaded", "edition", edition.EditionID, "md5", edition.MD5)
		return nil
	}

	jobProcessor := internal.NewJobProcessor(ctx, c.parallelism)
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

	if c.printOutput {
		result, err := makeOutput(allEditions, oldEditionsHash)
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
	exp.MaxElapsedTime = c.retryFor
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
			slog.Debug(logMsg, "retrying-in", d, "error", err)
		},
	)
}
