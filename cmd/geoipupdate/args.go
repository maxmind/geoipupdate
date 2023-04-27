package main

import (
	"errors"

	"github.com/maxmind/geoipupdate/v5/pkg/geoipupdate/vars"
	flag "github.com/spf13/pflag"
)

// Args are command line arguments.
type Args struct {
	ConfigFile        string
	DatabaseDirectory string
	StackTrace        bool
	Verbose           bool
	DisplayVersion    bool
	Parallelism       int
}

func getArgs() (*Args, error) {
	configFile := flag.StringP(
		"config-file",
		"f",
		vars.DefaultConfigFile,
		"Configuration file",
	)
	databaseDirectory := flag.StringP(
		"database-directory",
		"d",
		"",
		"Store databases in this directory (uses config if not specified)",
	)
	stackTrace := flag.Bool("stack-trace", false, "Show a stack trace along with any error message.")
	verbose := flag.BoolP("verbose", "v", false, "Use verbose output")
	displayVersion := flag.BoolP("version", "V", false, "Display the version and exit")
	parallelism := flag.Int("parallelism", 0, "Set the number of parallel database downloads.")
	flag.Parse()

	if *configFile == "" {
		return nil, errors.New("You must provide a configuration file.")
	}

	if *parallelism < 0 {
		return nil, errors.New("Parallelism must be a positive number")
	}

	return &Args{
		ConfigFile:        *configFile,
		DatabaseDirectory: *databaseDirectory,
		StackTrace:        *stackTrace,
		Verbose:           *verbose,
		DisplayVersion:    *displayVersion,
		Parallelism:       *parallelism,
	}, nil
}
