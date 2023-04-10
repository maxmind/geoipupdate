// Package geoipupdate provides a library for using MaxMind's GeoIP Update
// service.
package geoipupdate

import (
	"context"
	"fmt"
	"log"

	"github.com/maxmind/geoipupdate/v4/pkg/geoipupdate/database"
	"github.com/maxmind/geoipupdate/v4/pkg/geoipupdate/internal"
)

// Client uses config data to initiate a download or update
// process for GeoIP databases.
type Client struct {
	config *Config
}

// NewClient initialized a new Client struct.
func NewClient(config *Config) *Client {
	return &Client{config: config}
}

// Run starts the download or update process.
func (c *Client) Run(ctx context.Context) error {
	fileLock, err := internal.NewFileLock(c.config.LockFile, c.config.Verbose)
	if err != nil {
		return fmt.Errorf("error initializing file lock: %w", err)
	}
	if err := fileLock.Acquire(); err != nil {
		return fmt.Errorf("error acquiring file lock: %w", err)
	}
	defer func() {
		if err := fileLock.Release(); err != nil {
			log.Printf("error releasing file lock: %s", err)
		}
	}()

	jobProcessor := internal.NewJobProcessor(ctx, c.config.Parallelism)

	reader := database.NewHTTPReader(
		c.config.Proxy,
		c.config.URL,
		c.config.AccountID,
		c.config.LicenseKey,
		c.config.RetryFor,
		c.config.Verbose,
	)

	writer, err := database.NewLocalFileWriter(
		c.config.DatabaseDirectory,
		c.config.PreserveFileTimes,
		c.config.Verbose,
	)
	if err != nil {
		return fmt.Errorf("error initializing database writer: %w", err)
	}

	for _, editionID := range c.config.EditionIDs {
		editionID := editionID
		processFunc := func(ctx context.Context) error {
			editionHash, err := writer.GetHash(editionID)
			if err != nil {
				return err
			}

			result, err := reader.Read(ctx, editionID, editionHash)
			if err != nil {
				return err
			}

			if err := writer.Write(result); err != nil {
				return err
			}

			return nil
		}

		jobProcessor.Add(processFunc)
	}

	// Run blocks until all jobs are processed or exits early after
	// the first encountered error.
	if err := jobProcessor.Run(ctx); err != nil {
		return fmt.Errorf("error running the job processor: %w", err)
	}

	return nil
}
