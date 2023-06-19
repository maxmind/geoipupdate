// geoipupdate performs automatic updates of GeoIP binary databases.
package main

import (
	"context"
	"fmt"
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
		geoipupdate.WithDatabaseDirectory(args.DatabaseDirectory),
		geoipupdate.WithParallelism(args.Parallelism),
		geoipupdate.WithVerbose(args.Verbose),
		geoipupdate.WithOutput(args.Output),
	)
	if err != nil {
		fatalLogger(fmt.Sprintf("Error loading configuration file %s", args.ConfigFile), err)
	}

	if config.Verbose {
		log.Printf("geoipupdate version %s", version)
		log.Printf("Using config file %s", args.ConfigFile)
		log.Printf("Using database directory %s", config.DatabaseDirectory)
	}

	client := geoipupdate.NewClient(config)
	if err = client.Run(context.Background()); err != nil {
		fatalLogger("Error retrieving updates", err)
	}
}
