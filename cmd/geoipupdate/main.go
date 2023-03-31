// geoipupdate performs automatic updates of GeoIP binary databases.
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/maxmind/geoipupdate/v4/pkg/geoipupdate"
	"github.com/maxmind/geoipupdate/v4/pkg/geoipupdate/database"
	"golang.org/x/sync/errgroup"
)

var (
	version                  = "unknown"
	defaultConfigFile        string
	defaultDatabaseDirectory string
)

func main() {
	log.SetFlags(0)

	if defaultConfigFile == "" {
		defaultConfigFile = geoipupdate.DefaultConfigFile
	}
	if defaultDatabaseDirectory == "" {
		defaultDatabaseDirectory = geoipupdate.DefaultDatabaseDirectory
	}

	args := getArgs()
	fatalLogger := func(message string, err error) {
		if args.StackTrace {
			log.Printf("%s: %+v", message, err)
		} else {
			log.Printf("%s: %s", message, err)
		}
		os.Exit(1)
	}

	config, err := geoipupdate.NewConfig(
		args.ConfigFile,
		defaultDatabaseDirectory,
		args.DatabaseDirectory,
		args.Verbose,
		geoipupdate.WithParallelism(args.Parallelism),
	)
	if err != nil {
		fatalLogger(fmt.Sprintf("error loading configuration file %s", args.ConfigFile), err)
	}

	if config.Verbose {
		log.Printf("geoipupdate version %s", version)
		log.Printf("Using config file %s", args.ConfigFile)
		log.Printf("Using database directory %s", config.DatabaseDirectory)
	}

	client := geoipupdate.NewClient(config)

	downloadFunc := func(editionID string) error {
		return download(client, config, editionID)
	}

	if err = run(config, downloadFunc); err != nil {
		fatalLogger("error retrieving updates", err)
	}
}

// run concurrently downloads GeoIP databases using the provided config.
// `config.Parallelism` limits the number of concurrent downloads that can be
// executed and setting it to 1 would just mean databases would download
// sequencially.
func run(
	config *geoipupdate.Config,
	downloadFunc func(string) error,
) error {
	g := new(errgroup.Group)
	waitChan := make(chan struct{}, config.Parallelism)
	for _, editionID := range config.EditionIDs {
		waitChan <- struct{}{}
		editionID := editionID
		g.Go(func() error {
			defer func() { <-waitChan }()
			return downloadFunc(editionID)
		})
	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf("download error: %w", err)
	}
	return nil
}

// download fetches a specific database edition and writes it to a local file.
func download(
	client *http.Client,
	config *geoipupdate.Config,
	editionID string,
) error {
	filename, err := geoipupdate.GetFilename(config, editionID, client)
	if err != nil {
		return fmt.Errorf("error retrieving filename for %s: %w", editionID, err)
	}

	filePath := filepath.Join(config.DatabaseDirectory, filename)
	dbWriter, err := database.NewLocalFileDatabaseWriter(filePath, config.LockFile, config.Verbose)
	if err != nil {
		return fmt.Errorf("error creating database writer for %s: %w", editionID, err)
	}

	dbReader := database.NewHTTPDatabaseReader(client, config)
	if err := dbReader.Get(dbWriter, editionID); err != nil {
		return fmt.Errorf("error while getting database for %s: %w", editionID, err)
	}
	return nil
}
