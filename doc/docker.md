# Docker

## Configuring

The Docker image is configured by environment variables. The following
variables are required:

* `GEOIPUPDATE_ACCOUNT_ID` - Your MaxMind account ID.
* `GEOIPUPDATE_LICENSE_KEY` - Your case-sensitive MaxMind license key.
* `GEOIPUPDATE_EDITION_IDS` - List of space-separated database edition IDs.
  Edition IDs may consist of letters, digits, and dashes. For example,
  `GeoIP2-City 106` would download the GeoIP2 City database
  (`GeoIP2-City`) and the GeoIP Legacy Country database (`106`).

The following are optional:

* `GEOIPUPDATE_FREQUENCY` - The number of hours between `geoipupdate` runs.
  If this is not set or is set to `0`, `geoipupdate` will run once and exit.
* `GEOIPUPDATE_HOST` - The host name of the server to use. The default is
  `updates.maxmind.com`.
* `GEOIPUPDATE_PROXY` - The proxy host name or IP address. You may optionally
  specify a port number, e.g., 127.0.0.1:8888. If no port number is specified,
  1080 will be used.
* `GEOIPUPDATE_PROXY_USER_PASSWORD` - The proxy user name and password,
  separated by a colon. For instance, `username:password`.
* `GEOIPUPDATE_PRESERVE_FILE_TIMES` - Whether to preserve modification times
  of files downloaded from the server. This option is either `0` or `1`. The
  default is `0`.
* `GEOIPUPDATE_VERBOSE` - Enable verbose mode. Prints out the steps that
  `geoipupdate` takes.

The environment variables can be placed in a file with one per line and
passed in with the `--env-file` flag. Alternatively, you may pass them in
individually with the `-e` flag.

## Running

Run the latest image with:

```
docker run --env-file <file>  -v <database directory>:/usr/share/GeoIP maxmind/geoipupdate
```

`<file>` should be the environment variable file with your configuration.
`<database directory>` should be the local directory that you want to download
the databases to.
