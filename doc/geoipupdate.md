# NAME

geoipupdate - GeoIP and GeoLite Update Program

# SYNOPSIS

**geoipupdate** [-Vvh] [-f *CONFIG_FILE*] [-d *TARGET_DIRECTORY*]

# DESCRIPTION

`geoipupdate` automatically updates GeoIP and GeoLite databases. The
program connects to the MaxMind GeoIP Update server to check for new
databases. If a new database is available, the program will download and
install it.

If you are using a firewall, you must have the DNS and HTTPS ports
open.

# OPTIONS

`-d`, `--database-directory`

:   Install databases to a custom directory.  This is optional. If provided, it
    overrides the `DatabaseDirectory` value from the configuration file and the
    `GEOIPUPDATE_DB_DIR` environment variable.

`-f`, `--config-file`

:   The configuration file to use. See `GeoIP.conf` and its documentation for
    more information. This is optional. It defaults to the environment variable
    `GEOIPUPDATE_CONF_FILE` if it is set, or CONFFILE otherwise.

`--parallelism`

:	Set the number of parallel database downloads.

`-h`, `--help`

:   Display help and exit.

`-V`, `--version`

:   Display version information and exit.

`-v`, `--verbose`

:   Enable verbose mode. Prints out the steps that `geoipupdate` takes. If
    provided, it overrides any `GEOIPUPDATE_VERBOSE` environment variable.

`-o`, `--output`

:   Output download/update results in JSON format.

# EXIT STATUS

`geoipupdate` returns 0 on success and 1 on error.

# NOTES

Typically you should run `geoipupdate` at least twice a week. Consult
our
[database release schedule](https://support.maxmind.com/knowledge-base/articles/download-and-update-maxmind-databases)
for more information.

On most Unix-like systems, this can be achieved by using cron. You can
find
[an example crontab file on our Developer Portal](https://dev.maxmind.com/geoip/updating-databases/#3-run-geoip-update).

To use with a proxy server, update your `GeoIP.conf` file as specified in
the `GeoIP.conf` man page. Alternatively, set the `GEOIPUPDATE_PROXY` or
`http_proxy` environment variable.

# BUGS

Report bugs to [support@maxmind.com](mailto:support@maxmind.com).

# AUTHORS

Written by William Storey.

This software is Copyright (c) 2018-2026 by MaxMind, Inc.

This is free software, licensed under the Apache License, Version 2.0 or
the MIT License, at your option.

# MORE INFORMATION

Visit [our website](https://www.maxmind.com/en/solutions/ip-geolocation-databases-api-services)
to learn more about the GeoIP databases or to sign up for a subscription.

# SEE ALSO

`GeoIP.conf`(5)
