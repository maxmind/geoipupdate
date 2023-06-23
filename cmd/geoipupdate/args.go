package main

import (
	"log"
	"os"

	"github.com/maxmind/geoipupdate/v5/pkg/geoipupdate/vars"
	flag "github.com/spf13/pflag"
)

// Args are command line arguments.
type Args struct {
	ConfigFile        string
	DatabaseDirectory string
	StackTrace        bool
	Verbose           bool
	Output            bool
	Parallelism       int
}

func getArgs() *Args {
	confFileDefault := vars.DefaultConfigFile
	if value, ok := os.LookupEnv("GEOIPUPDATE_CONF_FILE"); ok {
		confFileDefault = value
	}

	configFile := flag.StringP(
		"config-file",
		"f",
		confFileDefault,
		"Configuration file",
	)
	databaseDirectory := flag.StringP(
		"database-directory",
		"d",
		"",
		"Store databases in this directory (uses config if not specified)",
	)
	help := flag.BoolP("help", "h", false, "Display help and exit")
	stackTrace := flag.Bool("stack-trace", false, "Show a stack trace along with any error message")
	verbose := flag.BoolP("verbose", "v", false, "Use verbose output")
	output := flag.BoolP("output", "o", false, "Output download/update results in JSON format")
	displayVersion := flag.BoolP("version", "V", false, "Display the version and exit")
	parallelism := flag.Int("parallelism", 0, "Set the number of parallel database downloads")

	flag.Parse()

	if *help {
		printUsage()
	}
	if *displayVersion {
		log.Printf("geoipupdate %s", version)
		//nolint: revive // deep exit from main package
		os.Exit(0)
	}

	if *configFile == "" {
		log.Printf("You must provide a configuration file.")
		printUsage()
	}

	if *parallelism < 0 {
		log.Printf("Parallelism must be a positive number")
		printUsage()
	}

	return &Args{
		ConfigFile:        *configFile,
		DatabaseDirectory: *databaseDirectory,
		StackTrace:        *stackTrace,
		Verbose:           *verbose,
		Output:            *output,
		Parallelism:       *parallelism,
	}
}

func printUsage() {
	log.Printf("Usage: %s <arguments>\n", os.Args[0])
	flag.PrintDefaults()
	//nolint: revive // deep exit from main package
	os.Exit(1)
}
