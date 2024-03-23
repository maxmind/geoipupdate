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

	"github.com/cenkalti/backoff/v4"

	"github.com/maxmind/geoipupdate/v6/internal"
	"github.com/maxmind/geoipupdate/v6/internal/geoipupdate/database"
)

// Updater uses config data to initiate a download or update
// process for GeoIP databases.
type Updater struct {
	config *Config
	reader database.Reader
	output *log.Logger
	writer database.Writer
}

// NewUpdater initialized a new Updater struct.
func NewUpdater(config *Config) (*Updater, error) {
	reader := database.NewHTTPReader(
		config.Proxy,
		config.URL,
		config.AccountID,
		config.LicenseKey,
		config.Verbose,
	)

	writer, err := database.NewLocalFileWriter(
		config.DatabaseDirectory,
		config.PreserveFileTimes,
		config.Verbose,
	)
	if err != nil {
		return nil, err
	}

	return &Updater{
		config: config,
		reader: reader,
		output: log.New(os.Stdout, "", 0),
		writer: writer,
	}, nil
}

// Run starts the download or update process.
func (u *Updater) Run(ctx context.Context) error {
	fileLock, err := internal.NewFileLock(u.config.LockFile, u.config.Verbose)
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

	jobProcessor := internal.NewJobProcessor(ctx, u.config.Parallelism)

	var editions []database.ReadResult
	var mu sync.Mutex
	for _, editionID := range u.config.EditionIDs {
		editionID := editionID
		processFunc := func(ctx context.Context) error {
			edition, err := u.downloadEdition(ctx, editionID, u.reader, u.writer)
			if err != nil {
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

	if u.config.Output {
		result, err := json.Marshal(editions)
		if err != nil {
			return fmt.Errorf("marshaling result log: %w", err)
		}
		u.output.Print(string(result))
	}

	return nil
}

// downloadEdition downloads the file with retries.
func (u *Updater) downloadEdition(
	ctx context.Context,
	editionID string,
	r database.Reader,
	w database.Writer,
) (*database.ReadResult, error) {
	editionHash, err := w.GetHash(editionID)
	if err != nil {
		return nil, err
	}

	// RetryFor value of 0 means that no retries should be performed.
	// Max zero retries has to be set to achieve that
	// because the backoff never stops if MaxElapsedTime is zero.
	exp := backoff.NewExponentialBackOff()
	exp.MaxElapsedTime = u.config.RetryFor
	b := backoff.BackOff(exp)
	if exp.MaxElapsedTime == 0 {
		b = backoff.WithMaxRetries(exp, 0)
	}

	var edition *database.ReadResult
	err = backoff.RetryNotify(
		func() error {
			if edition, err = r.Read(ctx, editionID, editionHash); err != nil {
				if internal.IsPermanentError(err) {
					return backoff.Permanent(err)
				}

				return err
			}

			if err = w.Write(edition); err != nil {
				if internal.IsPermanentError(err) {
					return backoff.Permanent(err)
				}

				return err
			}

			return nil
		},
		b,
		func(err error, d time.Duration) {
			if u.config.Verbose {
				log.Printf("Couldn't download %s, retrying in %v: %v", editionID, d, err)
			}
		},
	)
	if err != nil {
		return nil, err
	}

	return edition, nil
}
