# GeoIP Update

The GeoIP Update program performs automatic updates of GeoIP2 and
GeoLite2 binary databases. CSV databases are _not_ supported.

## Installation

We provide releases for Linux, macOS (darwin), and Windows. Please see the
[Releases](https://github.com/maxmind/geoipupdate/releases) tab for the
latest release.

After you install GeoIP Update, please refer to our
[documentation](https://dev.maxmind.com/geoip/updating-databases?lang=en) for information
about configuration.

If you're upgrading from GeoIP Update 3.x, please see our [upgrade
guide](https://dev.maxmind.com/geoip/upgrading-geoip-update?lang=en).

### Installing on Linux via the tarball

Download and extract the appropriate tarball for your system. You will end
up with a directory named something like `geoipupdate_5.0.0_linux_amd64`
depending on the version and architecture.

Copy `geoipupdate` to where you want it to live. To install it to
`/usr/local/bin/geoipupdate`, run the equivalent of `sudo cp
geoipupdate_5.0.0_linux_amd64/geoipupdate /usr/local/bin`.

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

Run `dpkg -i path/to/geoipupdate_5.0.0_linux_amd64.deb` (replacing the
version number and architecture as necessary). You will need to be root.
For Ubuntu you can prefix the command with `sudo`. This will install
`geoipupdate` to `/usr/bin/geoipupdate`.

`geoipupdate` looks for the config file `/etc/GeoIP.conf` by default.

### Installing on RedHat or CentOS via the rpm

You can also use the tarball.

Download the appropriate .rpm for your system.

Run `rpm -Uvhi path/to/geoipupdate_5.0.0_linux_amd64.rpm` (replacing the
version number and architecture as necessary). You will need to be root.
This will install `geoipupdate` to `/usr/bin/geoipupdate`.

`geoipupdate` looks for the config file `/etc/GeoIP.conf` by default.

### Installing on macOS (darwin) via the tarball

This is the same as installing on Linux via the tarball, except choose a
tarball with "darwin" in the name.

### Installing on macOS via Homebrew

If you are on macOS and you have [Homebrew](https://brew.sh/) you can install
`geoipupdate` via `brew`

```
$ brew install geoipupdate
```

### Installing on Windows

Download and extract the appropriate zip for your system. You will end up
with a directory named something like `geoipupdate_5.0.0_windows_amd64`
depending on the version and architecture.

Copy `geoipupdate.exe` to where you want it to live.

`geoipupdate` looks for the config file
`\ProgramData\MaxMind\GeoIPUpdate\GeoIP.conf` on your system drive by
default.

### Installing via Docker

Please see our [Docker documentation](doc/docker.md).

### Installation from source or Git

You need the Go compiler (1.21+). You can get it at the [Go
website](https://golang.org).

The easiest way is via `go install`:

    $ go install github.com/maxmind/geoipupdate/v7/cmd/geoipupdate@latest

This installs `geoipupdate` to `$GOPATH/bin/geoipupdate`.

# Configuring

Please see our [online guide](https://dev.maxmind.com/geoip/updating-databases?lang=en) for
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

Please report bugs by filing an issue with [our GitHub issue
tracker](https://github.com/maxmind/geoipupdate/issues).

# Copyright and License

This software is Copyright (c) 2018 - 2024 by MaxMind, Inc.

This is free software, licensed under the [Apache License, Version
2.0](LICENSE-APACHE) or the [MIT License](LICENSE-MIT), at your option.
