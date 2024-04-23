# NAME

GeoIP.conf - Configuration file for geoipupdate

# SYNOPSIS

This file allows you to configure your `geoipupdate` program to
download GeoIP2 and GeoLite2 databases.

# DESCRIPTION

The file consists of one setting per line. Lines starting with `#`
are comments and will not be processed. All setting keywords are case
sensitive.

## Required settings:

`AccountID`

:   Your MaxMind account ID. This was formerly known as `UserId`. This can be
    overridden at run time by either the `GEOIPUPDATE_ACCOUNT_ID` or the
    `GEOIPUPDATE_ACCOUNT_ID_FILE` environment variables.

`LicenseKey`

:   Your case-sensitive MaxMind license key. This can be overridden at run time
    by either the `GEOIPUPDATE_LICENSE_KEY` or `GEOIPUPDATE_LICENSE_KEY_FILE`
    environment variables.

`EditionIDs`

:   List of space-separated database edition IDs. Edition IDs may consist
    of letters, digits, and dashes.  For example, `GeoIP2-City` would
    download the GeoIP2 City database (`GeoIP2-City`). This can be overridden
    at run time by the `GEOIPUPDATE_EDITION_IDS` environment variable. Note:
    this was formerly called `ProductIds`.

## Optional settings:

`DatabaseDirectory`

:   The directory to store the database files. If not set, the default is
    DATADIR. This can be overridden at run time by the `GEOIPUPDATE_DB_DIR`
    environment variable or the `-d` command line argument.

`Host`

:   The host name of the server to use. The default is `https://updates.maxmind.com`.
    This can be overridden at run time by the `GEOIPUPDATE_HOST` environment
    variable.

`Proxy`

:   The proxy host name or IP address. You may optionally specify a port
    number, e.g., `127.0.0.1:8888`. If no port number is specified, 1080
    will be used. This can be overridden at run time by the
    `GEOIPUPDATE_PROXY` environment variable.

`ProxyUserPassword`

:   The proxy user name and password, separated by a colon. For instance,
    `username:password`. This can be overridden at run time by the
    `GEOIPUPDATE_PROXY_USER_PASSWORD` environment variable.

`PreserveFileTimes`

:   Whether to preserve modification times of files downloaded from the
    server. This option is either `0` or `1`. The default is `0`. This
    can be overridden at run time by the `GEOIPUPDATE_PRESERVE_FILE_TIMES`
    environment variable.

`LockFile`

:   The lock file to use. This ensures only one `geoipupdate` process can run
    at a time. Note: Once created, this lockfile is not removed from the
    filesystem. The default is `.geoipupdate.lock` under the
    `DatabaseDirectory`. This can be overridden at run time by the
    `GEOIPUPDATE_LOCK_FILE` environment variable.

`RetryFor`

:   The amount of time to retry for when errors during HTTP transactions are
    encountered. It can be specified as a (possibly fractional) decimal number
    followed by a unit suffix. Valid time units are `ns`, `us` (or `µs`), `ms`,
    `s`, `m`, `h`. The default is `5m` (5 minutes). This can be overridden at
    run time by the `GEOIPUPDATE_RETRY_FOR` environment variable.

`Parallelism`

:   The maximum number of parallel database downloads. The default is
    1, which means that databases will be downloaded sequentially. This can be
    overridden at run time by the `GEOIPUPDATE_PARALLELISM` environment
    variable or the `--parallelism` command line argument.

## Deprecated settings:

The following are deprecated and will be ignored if present:

`Protocol`

`SkipPeerVerification`

`SkipHostnameVerification`

# SEE ALSO

`geoipupdate`(1)
