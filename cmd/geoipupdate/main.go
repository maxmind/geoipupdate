// geoipupdate performs automatic updates of GeoIP binary databases.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/maxmind/geoipupdate/v4/pkg/geoipupdate"
	"github.com/maxmind/geoipupdate/v4/pkg/geoipupdate/vars"
)

var (
	version                  = "unknown"
	defaultConfigFile        string
	defaultDatabaseDirectory string
)

func main() {
	log.SetFlags(0)

	if defaultConfigFile == "" {
		defaultConfigFile = vars.DefaultConfigFile
	}
	if defaultDatabaseDirectory == "" {
		defaultDatabaseDirectory = vars.DefaultDatabaseDirectory
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
	if err = client.Run(context.Background()); err != nil {
		fatalLogger("error retrieving updates", err)
	}
}
