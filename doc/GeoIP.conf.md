# GeoIP.conf

## NAME

GeoIP.conf - Configuration file for geoipupdate

## SYNOPSIS

This file allows you to configure your `geoipupdate` program to
download GeoIP2, GeoLite2, and GeoIP Legacy databases.

## DESCRIPTION

The file consists of one setting per line. Lines starting with `#`
are comments and will not be processed. All setting keywords are case
sensitive.

### Required settings:

* `EditionIDs` - List of database edition IDs. Edition IDs may consist
  of letters, digits, and dashes (e.g., "GeoIP2-City", "106"). Note: this
  was formerly called `ProductIds`.

### Optional settings:

* `AccountID` - Your MaxMind account ID. This was formerly known as
  `UserId`.
* `DatabaseDirectory` - The directory to store the database files. If not
  set, the default is DATADIR. This can be overridden at run time by the
  `-d` command line argument.
* `Host` - The host name of the server to use. The default is
  `updates.maxmind.com`.
* `Proxy` - The proxy host name or IP address. You may optionally specify
  a port number, e.g., `127.0.0.1:8888`. If no port number is specified,
  1080 will be used.
* `ProxyUserPassword` - The proxy user name and password, separated by a
  colon. For instance, `username:password`.
* `PreserveFileTimes` - Whether to preserve modification times of files
  downloaded from the server. This option is either `0` or `1`. The default
  is `0`.
* `LicenseKey` - Your case-sensitive MaxMind license key.
* `LockFile` - The lock file to use. This ensures only one `geoipupdate`
  process can run at a time. Note: Once created, this lockfile is not removed
  from the filesystem. The default is `.geoipupdate.lock` under the
  `DatabaseDirectory`.

### Deprecated settings:

The following are deprecated and will be ignored if present:

* `Protocol`
* `SkipPeerVerification`
* `SkipHostnameVerification`

## FILES

* [`GeoIP.conf`](../conf/GeoIP.conf.default) - Default `geoipupdate`
  configuration file.

## SEE ALSO

[geoipupdate](geoipupdate.md)
