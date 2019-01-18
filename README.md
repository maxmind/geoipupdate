# GeoIP Update

[![Build Status](https://travis-ci.com/maxmind/geoipupdate.svg?branch=master)](https://travis-ci.com/maxmind/geoipupdate)

The GeoIP Update program performs automatic updates of GeoIP2 and GeoIP Legacy
binary databases. CSV databases are _not_ supported.

This is the new version of GeoIP Update. If for some reason you need the
legacy C version, you can find it
[here](https://github.com/maxmind/geoipupdate-legacy).

## Installation

We provide releases for Linux, macOS (darwin), and Windows. Please see the
[Releases](https://github.com/maxmind/geoipupdate/releases) tab for the
latest release.

After you install geoipupdate, please refer to our
[documentation](https://dev.maxmind.com/geoip/geoipupdate/) for information
about configuration.

If you're upgrading from geoipupdate 3.x, please see our [upgrade
guide](https://dev.maxmind.com/geoip/geoipupdate/upgrading-to-geoip-update-4-x/).

### Installing on Linux via the tarball

Download and extract the appropriate tarball for your system. You will end
up with a directory named something like `geoipupdate_4.0.0_linux_amd64`
depending on the version and architecture.

Copy `geoipupdate` to where you want it to live. To install it to
`/usr/local/bin/geoipupdate`, run the equivalent of `sudo cp
geoipupdate_4.0.0_linux_amd64/geoipupdate /usr/local/bin`.

`geoipupdate` looks for the config file `/usr/local/etc/GeoIP.conf` by
default.

### Installing on Ubuntu via PPA

MaxMind provides a PPA for recent versions of Ubuntu. To add the PPA to
your sources, run:

```
$ sudo add-apt-repository ppa:maxmind/ppa
```

Then install `geoipupdate` by running:

```
$ sudo apt update
$ sudo apt install geoipupdate
```

### Installing on Ubuntu or Debian via the deb

You can also use the tarball.

Download the appropriate .deb for your system.

Run `dpkg -i path/to/geoipupdate_4.0.0_linux_amd64.deb` (replacing the
version number and architecture as necessary). You will need to be root.
For Ubuntu you can prefix the command with `sudo`. This will install
`geoipupdate` to `/usr/bin/geoipupdate`.

`geoipupdate` looks for the config file `/etc/GeoIP.conf` by default.

### Installing on RedHat or CentOS via the rpm

You can also use the tarball.

Download the appropriate .rpm for your system.

Run `rpm -i path/to/geoipupdate_4.0.0_linux_amd64.rpm` (replacing the
version number and architecture as necessary). You will need to be root.
This will install `geoipupdate` to `/usr/bin/geoipupdate`.

`geoipupdate` looks for the config file `/etc/GeoIP.conf` by default.

### Installing on macOS (darwin) via the tarball

This is the same as installing on Linux via the tarball, except choose a
tarball with "darwin" in the name.

### Installing on macOS via Homebrew

If you are on macOS and you have [Homebrew](http://brew.sh/) you can install
`geoipupdate` via `brew`

```
$ brew install geoipupdate
```

### Installing on Windows

Download and extract the appropriate zip for your system. You will end up
with a directory named something like `geoipupdate_4.0.0_windows_amd64`
depending on the version and architecture.

Copy `geoipupdate.exe` to where you want it to live.

`geoipupdate` looks for the config file
`\ProgramData\MaxMind/GeoIPUpdate\GeoIP.conf` on your system drive by
default.

### Installation from source or Git

You need the Go compiler (1.8+). You can get it at the [Go
website](https://golang.org).

The easiest way is via `go get`:

    $ go get -u github.com/maxmind/geoipupdate/cmd/geoipupdate

This installs `geoipupdate` to `$GOPATH/bin/geoipupdate`.

# Configuring

Please see our [online guide](https://dev.maxmind.com/geoip/geoipupdate/) for
directions on how to configure GeoIP Update.

# Documentation

See our documentation for the [`geoipupdate` program](doc/geoipupdate.md)
and the [`GeoIP.conf` configuration file](doc/GeoIP.conf.md).

# Default config file and database directory paths

We define default paths for the config file and database directory. If
these defaults are not appropriate for you, you can change them at build
time using flags:

    go build -ldflags "-X main.defaultConfigFile=/etc/GeoIP.conf \
        -X main.defaultDatabaseDirectory=/usr/share/GeoIP"

# Bug Reports

Please report bugs by filing an issue with our GitHub issue tracker at
https://github.com/maxmind/geoipupdate/issues

# Copyright and License

This software is Copyright (c) 2018 - 2019 by MaxMind, Inc.

This is free software, licensed under the [Apache License, Version
2.0](LICENSE-APACHE) or the [MIT License](LICENSE-MIT), at your option.
