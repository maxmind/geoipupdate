# GeoIP Update

## Description

The GeoIP Update program performs automatic updates of GeoIP2 and GeoIP Legacy
binary databases. CSV databases are _not_ supported.

Currently the program only supports Linux and other Unix-like systems.

## License

This library is licensed under the GNU General Public License version 2.

## Installing on Ubuntu

MaxMind provides a PPA for recent version of Ubuntu. To add the PPA to your
sources, run:

    $ sudo add-apt-repository ppa:maxmind/ppa

Then install `geoipupdate` by running:

    $ sudo aptitude update
    $ sudo aptitude install geoipupdate

## Installing From Source File

To install this from the source package, you will need a C compiler, Make,
and the curl library and headers. On Debian or Ubuntu, you can install these
dependencies by running:

    $ sudo apt-get install build-essential libcurl4-openssl-dev

Once you have the necessary dependencies, run the following commands:

    $ ./configure
    $ make
    $ sudo make install

The `configure` script takes the standard options to set where files are
installed such as `--prefix`, etc. See `./configure --help` for details.

## Installing From GitHub

To install from Git, you will need automake, autoconf, and libtool installed.

Our public git repository is hosted on GitHub at
https://github.com/maxmind/geoipupdate

You can clone this repository and bootstrap it by running:

    $ git clone https://github.com/maxmind/geoipupdate
    $ cd geoipupdate
    $ ./bootstrap

Then follow the instructions above for "Installing From Source Files".

# Configuring

Please see our [online guide](http://dev.maxmind.com/geoip/geoipupdate/) for
directions on how to configure GeoIP Update.

# Bug Reports

Please report bugs by filing an issue with our GitHub issue tracker at
https://github.com/maxmind/geoipupdate/issues

# License

This software is licensed under the GNU General Public License (GPL), version
2 or later.

# Copyright

This software is Copyright (c) 2014 - 2017 by MaxMind, Inc.
