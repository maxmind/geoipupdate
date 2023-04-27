// Package geoipupdate provides a library for using MaxMind's GeoIP Update
// service.
package geoipupdate

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/maxmind/geoipupdate/v5/pkg/geoipupdate/database"
	"github.com/maxmind/geoipupdate/v5/pkg/geoipupdate/internal"
	"github.com/maxmind/geoipupdate/v5/pkg/geoipupdate/vars"
)

var (
	output = vars.NewBareDiscardLogger()
	log    = vars.NewDiscardLogger("updater")
)

// Client uses config data to initiate a download or update
// process for GeoIP databases.
type Client struct {
	config    *Config
	getReader func() (database.Reader, error)
	getWriter func() (database.Writer, error)
}

// NewClient initialized a new Client struct.
func NewClient(config *Config) *Client {
	if config.Verbose {
		output.SetOutput(os.Stdout)
		log.SetOutput(os.Stderr)
	}

	getReader := func() (database.Reader, error) {
		return database.NewHTTPReader(
			config.Proxy,
			config.URL,
			config.AccountID,
			config.LicenseKey,
			config.RetryFor,
			config.Verbose,
		), nil
	}

	getWriter := func() (database.Writer, error) {
		return database.NewLocalFileWriter(
			config.DatabaseDirectory,
			config.PreserveFileTimes,
			config.Verbose,
		)
	}

	return &Client{
		config:    config,
		getReader: getReader,
		getWriter: getWriter,
	}
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

	reader, err := c.getReader()
	if err != nil {
		return fmt.Errorf("error initializing database reader: %w", err)
	}

	writer, err := c.getWriter()
	if err != nil {
		return fmt.Errorf("error initializing database writer: %w", err)
	}

	var editions []database.ReadResult
	var mu sync.Mutex
	for _, editionID := range c.config.EditionIDs {
		editionID := editionID
		processFunc := func(ctx context.Context) error {
			editionHash, err := writer.GetHash(editionID)
			if err != nil {
				return err
			}

			edition, err := reader.Read(ctx, editionID, editionHash)
			if err != nil {
				return err
			}

			if err := writer.Write(edition); err != nil {
				return err
			}
			edition.CheckedAt = time.Now()

			mu.Lock()
			editions = append(editions, *edition)
			mu.Unlock()
			return nil
		}

		jobProcessor.Add(processFunc)
	}

	// Run blocks until all jobs are processed or exits early after
	// the first encountered error.
	if err := jobProcessor.Run(ctx); err != nil {
		return fmt.Errorf("error running the job processor: %w", err)
	}

	result, err := json.Marshal(editions)
	if err != nil {
		return fmt.Errorf("error marshaling result log: %w", err)
	}
	output.Print(string(result))
	return nil
}
