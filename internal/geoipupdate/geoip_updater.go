// Package geoipupdate provides a library for using MaxMind's GeoIP Update
// service.
package geoipupdate

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"

	"github.com/maxmind/geoipupdate/v7/client"
	"github.com/maxmind/geoipupdate/v7/internal"
	"github.com/maxmind/geoipupdate/v7/internal/geoipupdate/database"
)

type updateClient interface {
	Download(context.Context, string, string) (client.DownloadResponse, error)
}

// Updater uses config data to initiate a download or update
// process for GeoIP databases.
type Updater struct {
	config       *Config
	output       *log.Logger
	updateClient updateClient
	writer       database.Writer
}

// NewUpdater initialized a new Updater struct.
func NewUpdater(config *Config) (*Updater, error) {
	transport := http.DefaultTransport
	if config.Proxy != nil {
		proxyFunc := http.ProxyURL(config.Proxy)
		transport.(*http.Transport).Proxy = proxyFunc
	}
	httpClient := &http.Client{Transport: transport}

	updateClient, err := client.New(
		config.AccountID,
		config.LicenseKey,
		client.WithEndpoint(config.URL),
		client.WithHTTPClient(httpClient),
	)
	if err != nil {
		return nil, err
	}

	writer, err := database.NewLocalFileWriter(
		config.DatabaseDirectory,
		config.PreserveFileTimes,
		config.Verbose,
	)
	if err != nil {
		return nil, err
	}

	return &Updater{
		config:       config,
		output:       log.New(os.Stdout, "", 0),
		updateClient: updateClient,
		writer:       writer,
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
			edition, err := u.downloadEdition(ctx, editionID, u.updateClient, u.writer)
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
	uc updateClient,
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
			res, err := uc.Download(ctx, editionID, editionHash)
			if err != nil {
				if internal.IsPermanentError(err) {
					return backoff.Permanent(err)
				}

				return err
			}
			defer res.Reader.Close()

			if !res.UpdateAvailable {
				if u.config.Verbose {
					log.Printf("No new updates available for %s", editionID)
					log.Printf("Database %s up to date", editionID)
				}

				edition = &database.ReadResult{
					EditionID: editionID,
					OldHash:   editionHash,
					NewHash:   editionHash,
				}
				return nil
			}

			if u.config.Verbose {
				log.Printf("Updates available for %s", editionID)
			}

			err = u.writer.Write(
				editionID,
				res.Reader,
				res.MD5,
				res.LastModified,
			)
			if err != nil {
				if internal.IsPermanentError(err) {
					return backoff.Permanent(err)
				}

				return err
			}

			edition = &database.ReadResult{
				EditionID:  editionID,
				OldHash:    editionHash,
				NewHash:    res.MD5,
				ModifiedAt: res.LastModified,
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
