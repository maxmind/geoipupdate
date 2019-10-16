package main

import (
	"fmt"
	"github.com/maxmind/geoipupdate/pkg/geoipupdate"
	"github.com/pkg/errors"
	"log"
	"os"
)

// version is the program's version number.
var version = "unknown"

func main() {
	log.SetFlags(0)

	args := getArgs()
	fatalLogger := func(err error) {
		if args.StackTrace {
			log.Print(fmt.Sprintf(": %+v", err))
		} else {
			log.Print(fmt.Sprintf(": %s", err))
		}
		os.Exit(1)
	}

	config, err := geoipupdate.NewConfig(args.ConfigFile, geoipupdate.DefaultDatabaseDirectory, args.DatabaseDirectory, args.Verbose)
	if err != nil {
		fatalLogger(errors.Wrap(err, "Error loading configuration file"))
	}
	if config.Verbose {
		log.Printf("Using config file %s", args.ConfigFile)
		log.Printf("Using database directory %s", config.DatabaseDirectory)
	}

	err = geoipupdate.Run(config)
	fatalLogger(err)
}
