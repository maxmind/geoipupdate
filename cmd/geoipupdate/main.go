// geoipupdate performs automatic updates of GeoIP binary databases.
package main

import (
	"context"
	"log"
	"os"

	"github.com/maxmind/geoipupdate/v5/pkg/geoipupdate"
	"github.com/maxmind/geoipupdate/v5/pkg/geoipupdate/vars"
)

// These values are set by build scripts. Changing the names of
// the variables should be considered a breaking change.
var (
	version                  = "unknown"
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
		log.Printf("error loading configuration: %s", err)
		os.Exit(1)
	}

	if config.Verbose {
		log.Printf("geoipupdate version %s", version)
		log.Printf("Using config file %s", args.ConfigFile)
		log.Printf("Using database directory %s", config.DatabaseDirectory)
	}

	client := geoipupdate.NewClient(config)
	if err = client.Run(context.Background()); err != nil {
		log.Printf("error retrieving updates: %s", err)
		os.Exit(1)
	}
}
