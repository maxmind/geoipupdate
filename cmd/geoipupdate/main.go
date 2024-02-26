// geoipupdate performs automatic updates of GeoIP binary databases.
package main

import (
	"context"
	"log"

	"github.com/maxmind/geoipupdate/v6/pkg/geoipupdate"
	"github.com/maxmind/geoipupdate/v6/pkg/geoipupdate/vars"
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

	config, err := geoipupdate.NewConfig(
		geoipupdate.WithConfigFile(args.ConfigFile),
		geoipupdate.WithDatabaseDirectory(args.DatabaseDirectory),
		geoipupdate.WithParallelism(args.Parallelism),
		geoipupdate.WithVerbose(args.Verbose),
		geoipupdate.WithOutput(args.Output),
	)
	if err != nil {
		log.Fatalf("Error loading configuration: %s", err)
	}

	if config.Verbose {
		log.Printf("geoipupdate version %s", version)
		log.Printf("Using config file %s", args.ConfigFile)
		log.Printf("Using database directory %s", config.DatabaseDirectory)
	}

	client, err := geoipupdate.NewClient(config)
	if err != nil {
		log.Fatalf("Error initializing download client: %s", err)
	}

	if err = client.Run(context.Background()); err != nil {
		log.Fatalf("Error retrieving updates: %s", err)
	}
}
