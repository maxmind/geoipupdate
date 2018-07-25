package main

import (
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
	flock "github.com/theckman/go-flock"
)

// Version is the program's version number.
const Version = "0.0.1"

func main() {
	log.SetFlags(0)

	args := getArgs()

	config, err := NewConfig(args.ConfigFile, args.DatabaseDirectory)
	if err != nil {
		fatal(args, "Error loading configuration file", err)
	}

	lock, err := setup(config, args.Verbose)
	if err != nil {
		fatal(args, "Error preparing to update", err)
	}
	defer func() {
		if err := lock.Unlock(); err != nil {
			fatal(args, "Error unlocking lock file", errors.Wrap(err, "unlocking"))
		}
	}()

	if err := run(config, args.Verbose); err != nil {
		fatal(args, "Error retrieving updates", err)
	}
}

func fatal(
	args *Args,
	msg string,
	err error,
) {
	if args.StackTrace {
		log.Print(msg + fmt.Sprintf(": %+v", err))
	} else {
		log.Print(msg + fmt.Sprintf(": %s", err))
	}
	os.Exit(1)
}

func setup(
	config *Config,
	verbose bool,
) (*flock.Flock, error) {
	if err := maybeSetProxy(config, verbose); err != nil {
		return nil, err
	}

	if err := checkEnvironment(config); err != nil {
		return nil, err
	}

	lock := flock.NewFlock(config.LockFile)
	ok, err := lock.TryLock()
	if err != nil {
		return nil, errors.Wrap(err, "error acquiring a lock")
	}
	if !ok {
		return nil, errors.Errorf("could not acquire lock on %s", config.LockFile)
	}
	if verbose {
		log.Printf("Acquired lock file lock (%s)", config.LockFile)
	}

	return lock, nil
}

// Do not set a timeout to allow for very slow connections. Note the client
// will have TCP KeepAlive's enabled by default due to using
// http.DefaultTransport (which uses a net.Dialer with KeepAlive set).
var client = &http.Client{}

func maybeSetProxy(
	config *Config,
	verbose bool,
) error {
	if config.Proxy == nil {
		return nil
	}

	if verbose {
		log.Printf("Using proxy: %s", config.Proxy)
	}
	http.DefaultTransport.(*http.Transport).Proxy = http.ProxyURL(config.Proxy)

	return nil
}

func checkEnvironment(
	config *Config,
) error {
	fi, err := os.Stat(config.DatabaseDirectory)
	if err != nil {
		return errors.Wrap(err, "database directory is not available")
	}

	if !fi.IsDir() {
		return errors.New("database directory is not a directory")
	}

	// I don't think there is a reliable cross platform way to check the
	// directory is writable. We'll discover that when we try to write to it
	// anyway.

	return nil
}

func run(
	config *Config,
	verbose bool,
) error {
	for _, editionID := range config.EditionIDs {
		if err := updateEdition(config, verbose, editionID); err != nil {
			return errors.WithMessage(err, "error updating "+editionID)
		}
	}
	return nil
}

func updateEdition(
	config *Config,
	verbose bool,
	editionID string,
) error {
	filename, err := getFilename(config, verbose, editionID)
	if err != nil {
		return errors.WithMessage(err, "error retrieving filename")
	}

	md5, err := getCurrentMD5(config, verbose, filename)
	if err != nil {
		return errors.WithMessage(err, "error retrieving current MD5 of "+filename)
	}

	if err := maybeUpdate(
		config,
		verbose,
		editionID,
		filename,
		md5,
	); err != nil {
		return errors.WithMessage(err, "error updating")
	}

	return nil
}

func getFilename(
	config *Config,
	verbose bool,
	editionID string,
) (string, error) {
	url := fmt.Sprintf(
		"%s/app/update_getfilename?product_id=%s",
		config.URL,
		url.QueryEscape(editionID),
	)

	if verbose {
		log.Printf("Performing get filename request to %s", url)
	}
	res, err := client.Get(url)
	if err != nil {
		return "", errors.Wrap(err, "error performing HTTP request")
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			log.Fatalf("Error closing response body: %+v", errors.Wrap(err, "closing body"))
		}
	}()

	buf, err := ioutil.ReadAll(io.LimitReader(res.Body, 256))
	if err != nil {
		return "", errors.Wrap(err, "error reading response body")
	}

	if res.StatusCode != http.StatusOK {
		return "", errors.Errorf("unexpected HTTP status code: %s: %s", res.Status, buf)
	}

	if len(buf) == 0 {
		return "", errors.New("response body is empty")
	}

	if bytes.Count(buf, []byte("\n")) > 0 ||
		bytes.Count(buf, []byte("\x00")) > 0 {
		return "", errors.New("invalid characters in filename")
	}

	return string(buf), nil
}

const zeroMD5 = "00000000000000000000000000000000"

func getCurrentMD5(
	config *Config,
	verbose bool,
	filename string,
) (string, error) {
	path := filepath.Join(config.DatabaseDirectory, filename)

	fh, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			if verbose {
				log.Printf("Not calculating MD5 sum as file does not exist: %s", path)
			}
			return zeroMD5, nil
		}
		return "", errors.Wrap(err, "error opening file")
	}
	defer func() {
		if err := fh.Close(); err != nil {
			log.Fatalf("Error closing file: %+v", errors.Wrap(err, "closing file"))
		}
	}()

	fi, err := fh.Stat()
	if err != nil {
		return "", errors.Wrap(err, "error stat'ing file")
	}
	if !fi.Mode().IsRegular() {
		return "", errors.New("not a regular file")
	}

	h := md5.New()
	if _, err := io.Copy(h, fh); err != nil {
		return "", errors.Wrap(err, "error reading file")
	}
	sum := fmt.Sprintf("%x", h.Sum(nil))
	if verbose {
		log.Printf("Calculated MD5 sum for %s: %s", path, sum)
	}
	return sum, nil
}

func maybeUpdate(
	config *Config,
	verbose bool,
	editionID,
	filename,
	md5 string,
) error {
	url := fmt.Sprintf(
		"%s/geoip/databases/%s/update?db_md5=%s",
		config.URL,
		url.PathEscape(editionID),
		url.QueryEscape(md5),
	)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return errors.Wrap(err, "error creating request")
	}
	if !strings.HasPrefix(editionID, "GeoLite2") {
		req.SetBasicAuth(fmt.Sprintf("%d", config.AccountID), config.LicenseKey)
	}

	if verbose {
		log.Printf("Performing update request to %s", url)
	}
	res, err := client.Do(req)
	if err != nil {
		return errors.Wrap(err, "error performing HTTP request")
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			log.Fatalf("Error closing response body: %+v", errors.Wrap(err, "closing body"))
		}
	}()

	if res.StatusCode == http.StatusNotModified {
		if verbose {
			log.Printf("No new updates available for %s", editionID)
		}
		return nil
	}

	if res.StatusCode != http.StatusOK {
		buf, err := ioutil.ReadAll(io.LimitReader(res.Body, 256))
		if err == nil {
			return errors.Errorf("unexpected HTTP status code: %s: %s", res.Status, buf)
		}
		return errors.Errorf("unexpected HTTP status code: %s", res.Status)
	}

	newMD5 := res.Header.Get("X-Database-MD5")
	if newMD5 == "" {
		return errors.New("no X-Database-MD5 header found")
	}
	lastModified, err := getLastModified(res.Header)
	if err != nil {
		return err
	}

	return writeAndCheck(config, verbose, filename, res.Body, newMD5, lastModified)
}

func getLastModified(
	headers http.Header,
) (time.Time, error) {
	lastModifiedStr := headers.Get("Last-Modified")
	if lastModifiedStr == "" {
		return time.Time{}, errors.New("no Last-Modified header found")
	}

	t, err := time.ParseInLocation(time.RFC1123, lastModifiedStr, time.UTC)
	if err != nil {
		return time.Time{}, errors.Wrap(err, "error parsing time")
	}

	return t, nil
}

func writeAndCheck(
	config *Config,
	verbose bool,
	filename string,
	body io.Reader,
	newMD5 string,
	lastModified time.Time,
) error {
	targetTest := filepath.Join(
		config.DatabaseDirectory,
		fmt.Sprintf("%s.test", filename),
	)

	fh, err := os.OpenFile(targetTest, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return errors.Wrap(err, "error creating file")
	}

	gzReader, err := gzip.NewReader(body)
	if err != nil {
		_ = fh.Close()
		_ = os.Remove(targetTest)
		return errors.Wrap(err, "error creating gzip reader")
	}

	md5Writer := md5.New()
	multiWriter := io.MultiWriter(fh, md5Writer)

	if _, err := io.Copy(multiWriter, gzReader); err != nil {
		_ = fh.Close()
		_ = os.Remove(targetTest)
		_ = gzReader.Close()
		return errors.Wrap(err, "error reading/writing")
	}

	if err := gzReader.Close(); err != nil {
		_ = fh.Close()
		_ = os.Remove(targetTest)
		return errors.Wrap(err, "error closing gzip reader")
	}

	if err := fh.Sync(); err != nil {
		_ = fh.Close()
		_ = os.Remove(targetTest)
		return errors.Wrap(err, "error syncing file")
	}

	if err := fh.Close(); err != nil {
		_ = os.Remove(targetTest)
		return errors.Wrap(err, "error closing file")
	}

	gotMD5 := fmt.Sprintf("%x", md5Writer.Sum(nil))
	if !strings.EqualFold(gotMD5, newMD5) {
		_ = os.Remove(targetTest)
		return errors.Errorf("MD5 of new database (%s) does not match expected MD5 (%s)",
			gotMD5, newMD5)
	}

	target := filepath.Join(config.DatabaseDirectory, filename)

	if err := os.Rename(targetTest, target); err != nil {
		_ = os.Remove(targetTest)
		return errors.New("error moving database into place")
	}

	if config.PreserveFileTimes {
		if err := os.Chtimes(target, lastModified, lastModified); err != nil {
			return errors.Wrap(err, "error setting times on file")
		}
	}

	// fsync the directory. http://austingroupbugs.net/view.php?id=672

	dh, err := os.Open(config.DatabaseDirectory)
	if err != nil {
		return errors.Wrap(err, "error opening database directory")
	}
	defer func() {
		if err := dh.Close(); err != nil {
			log.Fatalf("Error closing directory: %+v", errors.Wrap(err, "closing directory"))
		}
	}()

	if err := dh.Sync(); err != nil {
		return errors.Wrap(err, "error syncing database directory")
	}

	if verbose {
		log.Printf("Updated %s", target)
	}
	return nil
}
