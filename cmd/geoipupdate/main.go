// geoipupdate performs automatic updates of GeoIP binary databases.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/maxmind/geoipupdate/v5/pkg/geoipupdate"
	"github.com/maxmind/geoipupdate/v5/pkg/geoipupdate/vars"
	flag "github.com/spf13/pflag"
)

// These values are set by build scripts. Changing the names of
// the variables should be considered a breaking change.
var (
	version                  = "unknown"
	defaultConfigFile        string
	defaultDatabaseDirectory string
	log                      = vars.NewDiscardLogger("main")
)

func main() {
	if defaultConfigFile != "" {
		vars.DefaultConfigFile = defaultConfigFile
	}

	if defaultDatabaseDirectory != "" {
		vars.DefaultDatabaseDirectory = defaultDatabaseDirectory
	}

	args, err := getArgs()
	if err != nil {
		printUsage(err)
	}

	if args.DisplayVersion {
		exitWithCode(fmt.Sprintf("geoipupdate %s", version), 0)
	}

	if args.Verbose {
		log.SetOutput(os.Stderr)
	}

	fatalLogger := func(message string, err error) {
		format := "%s: %s"
		if args.StackTrace {
			format = "%s: %+v"
		}
		exitWithCode(fmt.Sprintf(format, message, err), 1)
	}

	config, err := geoipupdate.NewConfig(
		args.ConfigFile,
		geoipupdate.WithDatabaseDirectory(args.DatabaseDirectory),
		geoipupdate.WithParallelism(args.Parallelism),
		geoipupdate.WithVerbose(args.Verbose),
	)
	if err != nil {
		fatalLogger(fmt.Sprintf("error loading configuration file %s", args.ConfigFile), err)
	}

	log.Printf("geoipupdate version %s", version)
	log.Printf("Using config file %s", args.ConfigFile)
	log.Printf("Using database directory %s", config.DatabaseDirectory)

	client := geoipupdate.NewClient(config)
	if err = client.Run(context.Background()); err != nil {
		fatalLogger("error retrieving updates", err)
	}
}

// exitWithCode will print a messge to stderr and end the program with the
// specified exitCode.
func exitWithCode(msg string, exitCode int) {
	log := vars.NewBareStderrLogger()
	log.Print(msg)
	os.Exit(exitCode)
}

// printUsage prints the usage of the program along with the provided error.
func printUsage(err error) {
	log := vars.NewBareStderrLogger()
	if err != nil {
		log.Print(err.Error())
	}
	log.Printf("Usage: %s <arguments>\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(1)
}
