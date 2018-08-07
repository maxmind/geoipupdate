package main

import (
	"log"
	"os"

	flag "github.com/spf13/pflag"
)

// Args are command line arguments.
type Args struct {
	ConfigFile        string
	DatabaseDirectory string
	StackTrace        bool
	Verbose           bool
}

func getArgs() *Args {
	configFile := flag.StringP("config-file", "f", "", "Configuration file (required)")
	databaseDirectory := flag.StringP(
		"database-directory",
		"d",
		"",
		"Store databases in this directory (optional)",
	)
	help := flag.BoolP("help", "h", false, "Display help and exit")
	stackTrace := flag.Bool("stack-trace", false, "Show a stack trace along with any error message.")
	verbose := flag.BoolP("verbose", "v", false, "Use verbose output")
	version := flag.BoolP("version", "V", false, "Display the version and exit")

	flag.Parse()

	if *help {
		printUsage()
	}
	if *version {
		log.Printf("geoipupdate %s", Version)
		os.Exit(0)
	}

	if *configFile == "" {
		log.Printf("You must provide a configuration file.")
		printUsage()
	}

	return &Args{
		ConfigFile:        *configFile,
		DatabaseDirectory: *databaseDirectory,
		StackTrace:        *stackTrace,
		Verbose:           *verbose,
	}
}

func printUsage() {
	log.Printf("Usage: %s <arguments>\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(1)
}
