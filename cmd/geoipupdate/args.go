package main

import (
	"flag"
	"log"
	"os"
)

// Args are command line arguments.
type Args struct {
	ConfigFile        string
	DatabaseDirectory string
	StackTrace        bool
	Verbose           bool
}

func getArgs() *Args {
	configFile := flag.String("f", "", "Configuration file (required)")
	databaseDirectory := flag.String("d", "", "Store databases in this directory (optional)")
	help := flag.Bool("h", false, "Display help and exit")
	stackTrace := flag.Bool("stack-trace", false, "Show a stack trace along with any error message.")
	verbose := flag.Bool("v", false, "Use verbose output")
	version := flag.Bool("V", false, "Display the version and exit")

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
