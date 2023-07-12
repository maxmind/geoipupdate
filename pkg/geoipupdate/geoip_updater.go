// Package geoipupdate provides a library for using MaxMind's GeoIP Update
// service.
package geoipupdate

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/maxmind/geoipupdate/v6/pkg/geoipupdate/database"
	"github.com/maxmind/geoipupdate/v6/pkg/geoipupdate/internal"
)

// Client uses config data to initiate a download or update
// process for GeoIP databases.
type Client struct {
	config    *Config
	getReader func() (database.Reader, error)
	getWriter func() (database.Writer, error)
	output    *log.Logger
}

// NewClient initialized a new Client struct.
func NewClient(config *Config) *Client {
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
		output:    log.New(os.Stdout, "", 0),
	}
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

	jobProcessor := internal.NewJobProcessor(ctx, c.config.Parallelism)

	reader, err := c.getReader()
	if err != nil {
		return fmt.Errorf("initializing database reader: %w", err)
	}

	writer, err := c.getWriter()
	if err != nil {
		return fmt.Errorf("initializing database writer: %w", err)
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

			edition.CheckedAt = time.Now().In(time.UTC)

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
		return fmt.Errorf("running the job processor: %w", err)
	}

	if c.config.Output {
		result, err := json.Marshal(editions)
		if err != nil {
			return fmt.Errorf("marshaling result log: %w", err)
		}
		c.output.Print(string(result))
	}

	return nil
}
