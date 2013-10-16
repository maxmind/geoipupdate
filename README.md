# GeoIP Update

## Description

The GeoIP Update program performs automatic updates of GeoIP2 and GeoIP Legacy
binary databases. Currently the program only supports Linux and other Unix-
like systems.

## License

This library is licensed under the GNU General Public License version 3.

## Installing From Source File

To install this code, run the following commands:

    $ ./configure
    $ make
    $ sudo make install

The `configure` script takes the standard options to set where files are
installed such as `--prefix`, etc. See `./configure --help` for details.

## Installing From GitHub

Our public git repository is hosted on GitHub at
https://github.com/maxmind/geoipupdate

You can clone this repository and build it by running:

    $ git clone https://github.com/maxmind/geoipupdate
    $ ./bootstrap
    $ ./configure
    $ make
    $ make install

# Configuring

Please see our [online guide](http://dev.maxmind.com/geoip/geoipupdate/) for
directions on how to configure GeoIP Update.

# Bug Reports

Please report bugs by filing an issue with our GitHub issue tracker at
https://github.com/maxmind/geoipupdate/issues

# Copyright

This software is Copyright (c) 2013 by MaxMind, Inc.
