# GEOIPUPDATE

## NAME

geoipupdate - GeoIP2, GeoLite2, and GeoIP Legacy Update Program

## SYNOPSIS

geoipupdate \[-Vvh\] \[-f license\_file\] \[-d target\_directory\]

## DESCRIPTION

`geoipupdate` automatically updates GeoIP2, GeoLite2, and GeoIP Legacy
databases. The program connects to the MaxMind GeoIP Update server to
check for new databases. If a new database is available, the program will
download and install it.

If you are using a firewall, you must have the DNS and HTTPS ports
open.

## OPTIONS

* `-d` - Install databases to a custom directory. This must be specified
  if `DatabaseDirectory` is not set in the configuration file.
* `-f` - The configuration file to use. See [GeoIP.conf](GeoIP.conf.md)
  for more information. *Required*.
* `-h` - Display help and exit.
* `-stack-trace` - Show a stack trace on any error message. This is
  primarily useful for debugging.
* `-V` - Display version information and exit.
* `-v` - Enable verbose mode. Prints out the steps that `geoipupdate`
  takes.

## USAGE

Typically you should run `geoipupdate` weekly. One way to achieve this
is to use cron. Below is a sample crontab file that runs `geoipupdate`
on each Wednesday at noon:

```
# top of crontab

MAILTO=your@email.com

0 12 * * 3 BIN_DIR/geoipupdate

# end of crontab

```

To use with a proxy server, update your `GeoIP.conf` file as specified
in the `GeoIP.conf` man page or set the `http\_proxy` environment
variable.

## RETURN CODES

`geoipupdate` returns 0 on success and 1 on error.

## FILES

* `GeoIP.conf` - Configuration file for GeoIP Update. See the
  [`GeoIP.conf` documentation](GeoIP.conf.md) for more information.

## AUTHOR

Written by William Storey.

## REPORTING BUGS

Report bugs to [support@maxmind.com](mailto:support@maxmind.com).

## COPYRIGHT

This software is Copyright (c) 2018 by MaxMind, Inc.

This is free software, licensed under the [Apache License, Version
2.0](LICENSE-APACHE) or the [MIT License](LICENSE-MIT), at your option.

## MORE INFORMATION

Visit [our website](https://www.maxmind.com/en/geoip2-services-and-databases)
to learn more about the GeoIP2 and GeoIP Legacy databases or to sign up
for a subscription.

## SEE ALSO

[`GeoIP.conf`](GeoIP.conf.md)
