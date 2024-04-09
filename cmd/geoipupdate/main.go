// geoipupdate performs automatic updates of GeoIP binary databases.
package main

import (
	"context"
	"log"

	"github.com/maxmind/geoipupdate/v7/internal/geoipupdate"
	"github.com/maxmind/geoipupdate/v7/internal/vars"
)

const unknownVersion = "unknown"

// These values are set by build scripts. Changing the names of
// the variables should be considered a breaking change.
var (
	version                  = unknownVersion
	defaultConfigFile        string
	defaultDatabaseDirectory string
)

func main() {
	log.SetFlags(0)

	if defaultConfigFile != "" {
		vars.DefaultConfigFile = defaultConfigFile
	}

	if defaultDatabaseDirectory != "" {
		vars.DefaultDatabaseDirectory = defaultDatabaseDirectory
	}

	args := getArgs()

	opts := []geoipupdate.Option{
		geoipupdate.WithConfigFile(args.ConfigFile),
		geoipupdate.WithDatabaseDirectory(args.DatabaseDirectory),
		geoipupdate.WithParallelism(args.Parallelism),
	}

	if args.Output {
		opts = append(opts, geoipupdate.WithOutput)
	}

	if args.Verbose {
		opts = append(opts, geoipupdate.WithVerbose)
	}

	config, err := geoipupdate.NewConfig(opts...)
	if err != nil {
		log.Fatalf("Error loading configuration: %s", err)
	}

	if config.Verbose {
		log.Printf("geoipupdate version %s", version)
		log.Printf("Using config file %s", args.ConfigFile)
		log.Printf("Using database directory %s", config.DatabaseDirectory)
	}

	u, err := geoipupdate.NewUpdater(config)
	if err != nil {
		log.Fatalf("Error initializing updater: %s", err)
	}

	if err = u.Run(context.Background()); err != nil {
		log.Fatalf("Error retrieving updates: %s", err)
	}
}
