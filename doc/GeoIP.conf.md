# NAME

GeoIP.conf - Configuration file for geoipupdate

# SYNOPSIS

This file allows you to configure your `geoipupdate` program to
download GeoIP2, GeoLite2, and GeoIP Legacy databases.

# DESCRIPTION

The file consists of one setting per line. Lines starting with `#`
are comments and will not be processed. All setting keywords are case
sensitive.

## Required settings:

`AccountID`

:   Your MaxMind account ID. This was formerly known as `UserId`.

`LicenseKey`

:   Your case-sensitive MaxMind license key.

`EditionIDs`

:   List of space-separated database edition IDs. Edition IDs may consist
    of letters, digits, and dashes.  For example, `GeoIP2-City 106` would
    download the GeoIP2 City database (`GeoIP2-City`) and the GeoIP Legacy
    Country database (`106`). Note: this was formerly called `ProductIds`.

## Optional settings:

`DatabaseDirectory`

:   The directory to store the database files. If not set, the default is
    DATADIR. This can be overridden at run time by the `-d` command line
    argument.

`Host`

:   The host name of the server to use. The default is `updates.maxmind.com`.

`Proxy`

:   The proxy host name or IP address. You may optionally specify
    a port number, e.g., `127.0.0.1:8888`. If no port number is specified,
    1080 will be used.

`ProxyUserPassword`

:   The proxy user name and password, separated by a colon. For instance,
    `username:password`.

`PreserveFileTimes`

:   Whether to preserve modification times of files downloaded from the
    server. This option is either `0` or `1`. The default is `0`.

`LockFile`

:   The lock file to use. This ensures only one `geoipupdate` process can run
    at a time. Note: Once created, this lockfile is not removed from the
    filesystem. The default is `.geoipupdate.lock` under the
    `DatabaseDirectory`.

`RetryFor`

:   The amount of time to retry for when errors during HTTP transactions are
    encountered. It can be specified as a (possibly fractional) decimal number
    followed by a unit suffix. Valid time units are `ns`, `us` (or `µs`), `ms`,
    `s`, `m`, `h`. The default is `5m` (5 minutes).

## Deprecated settings:

The following are deprecated and will be ignored if present:

`Protocol`

`SkipPeerVerification`

`SkipHostnameVerification`

# SEE ALSO

`geoipupdate`(1)
