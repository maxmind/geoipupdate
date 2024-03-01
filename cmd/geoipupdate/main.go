// geoipupdate performs automatic updates of GeoIP binary databases.
package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/maxmind/geoipupdate/v6/pkg/geoipupdate"
	"github.com/maxmind/geoipupdate/v6/pkg/geoipupdate/api"
	"github.com/maxmind/geoipupdate/v6/pkg/geoipupdate/config"
	"github.com/maxmind/geoipupdate/v6/pkg/geoipupdate/lock"
	"github.com/maxmind/geoipupdate/v6/pkg/geoipupdate/vars"
	geoipupdatewriter "github.com/maxmind/geoipupdate/v6/pkg/geoipupdate/writer"
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

	conf, err := config.NewConfig(
		config.WithConfigFile(args.ConfigFile),
		config.WithDatabaseDirectory(args.DatabaseDirectory),
		config.WithParallelism(args.Parallelism),
		config.WithVerbose(args.Verbose),
		config.WithOutput(args.Output),
	)
	if err != nil {
		log.Fatalf("Error loading configuration: %s", err)
	}

	options := &slog.HandlerOptions{Level: slog.LevelInfo}
	if conf.Verbose {
		options = &slog.HandlerOptions{Level: slog.LevelDebug}
	}
	handler := slog.NewTextHandler(os.Stderr, options)
	slog.SetDefault(slog.New(handler))

	slog.Debug("geoipupdate", "version", version)
	slog.Debug("config file", "path", args.ConfigFile)
	slog.Debug("database directory", "path", conf.DatabaseDirectory)

	transport := http.DefaultTransport
	if conf.Proxy != nil {
		proxyFunc := http.ProxyURL(conf.Proxy)
		transport.(*http.Transport).Proxy = proxyFunc
	}

	downloader := api.NewHTTPDownloader(
		conf.AccountID,
		conf.LicenseKey,
		&http.Client{Transport: transport},
		conf.URL,
	)

	writer := geoipupdatewriter.NewDiskWriter(
		conf.DatabaseDirectory,
		conf.PreserveFileTimes,
	)

	slog.Debug("initializing file lock", "path", conf.LockFile)
	locker, err := lock.NewFileLock(conf.LockFile)
	if err != nil {
		slog.Error("setting up file lock", "error", err)
		os.Exit(1)
	}

	client := geoipupdate.NewClient(
		conf,
		downloader,
		locker,
		writer,
	)
	if err = client.Run(context.Background()); err != nil {
		slog.Error("retrieving updates", "error", err)
		os.Exit(1)
	}
}
