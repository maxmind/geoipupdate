# NAME

GeoIP.env - Configuration environment variables for geoipupdate

# SYNOPSIS

These environment variables allow you to configure your `geoipupdate`
program to download GeoIP2 and GeoLite2 databases.

# DESCRIPTION

Environment variables override any associated configuration that may
have been set from the config file.

## Required settings:

`GEOIPUPDATE_ACCOUNT_ID`

:   Your MaxMind account ID. This can be overriden at runtime by the
`--parallelism` command line argument.

`GEOIPUPDATE_LICENSE_KEY`

:   Your case-sensitive MaxMind license key.

`GEOIPUPDATE_EDITION_IDS`

:   List of space-separated database edition IDs. Edition IDs may consist
    of letters, digits, and dashes.  For example, `GeoIP2-City` would
    download the GeoIP2 City database (`GeoIP2-City`).

## Optional settings:

`GEOIPUPDATE_DB_DIR`

:   The directory to store the database files. If not set, the default is
    DATADIR. This can be overridden at run time by the `-d` command line
    argument.

`GEOIPUPDATE_HOST`

:   The host name of the server to use. The default is `updates.maxmind.com`.

`GEOUPUPDATE_PROXY`

:   The proxy host name or IP address. You may optionally specify
    a port number, e.g., `127.0.0.1:8888`. If no port number is specified,
    1080 will be used.

`GEOIPUPDATE_PROXY_USER_PASSWORD`

:   The proxy user name and password, separated by a colon. For instance,
    `username:password`.

`GEOIPUPDATE_PRESERVE_FILE_TIMES`

:   Whether to preserve modification times of files downloaded from the
    server. This option is either `0` or `1`. The default is `0`.

`GEOIPUPDATE_LOCK_FILE`

:   The lock file to use. This ensures only one `geoipupdate` process can run
    at a time. Note: Once created, this lockfile is not removed from the
    filesystem. The default is `.geoipupdate.lock` under the
    `GEOIPUPDATE_DB_DIR`.

`GEOIPUPDATE_RETRY_FOR`

:   The amount of time to retry for when errors during HTTP transactions are
    encountered. It can be specified as a (possibly fractional) decimal number
    followed by a unit suffix. Valid time units are `ns`, `us` (or `µs`), `ms`,
    `s`, `m`, `h`. The default is `5m` (5 minutes).

`GEOIPUPDATE_PARALLELISM`

:	The maximum number of parallel database downloads. The default is
	1, which means that databases will be downloaded sequentially. This can be
	overriden at runtime by the `--parallelism` command line argument.

`VERBOSE`

:	Enable verbose mode. Prints out the steps that `geoipupdate` takes. This
    can be overriden at runtime by the `--verbose` command line argument.

# SEE ALSO

`geoipupdate`(1)
`GeoIP.conf`(5)
