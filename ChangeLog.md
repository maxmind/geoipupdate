GeoIP Update Change Log
=======================

Unreleased

* `geoipupdate` now checks that the database directory is writable. If it
  is not, it reports the problem and aborts.
* `geoipupdate` now acquires a lock when starting up to ensure only one
  instance may run at a time. A new option, `LockFile`, exists to set the
  file to use as a lock. By default, `LockFile` is the file
  `.geoipupdate.lock` in the database directory.
* `geoipupdate` now prints out additional information from the server when
  a download request results in something other than HTTP status 2xx. This
  provides more information when the API does not respond with a database
  file. In conjunction with changes to the download service itself, errors
  such as lacking a subscription no longer show up with the message "not a
  valid gzip file".
* ${datarootdir}/GeoIP is now created on `make install`. Reported by Antonios
  Karagiannis. GitHub #29.
* Previously, a variable named `ERROR` was used. This caused issues building
  on Windows. Reported by Gisle Vanem. GitHub #36.

2.3.1 (2017-01-05)
------------------

* 2.3.0 was missing `GeoIP.conf.default`. This was added to the dist.
* The directory creation of `$(sysconfdir)` added in 2.3.0 incorrectly ran if
  the directory already existed rather than if it did not exist.

2.3.0 (2017-01-04)
------------------

* `geoipupdate` now uses TCP keep-alive when compiled with cURL 7.25 or
  greater.
* Previously, on an invalid gzip file, `geoipupdate` would output binary data
  to stderr. It now displays an appropriate error message.
* Install README, ChangeLog, GeoIP.conf.default etc into docdir. PR by
  Philip Prindeville. GitHub #33.
* `$(sysconfdir)` is now created if it doesn't exist. PR by Philip
  Prindeville. GitHub #33.
* The sample config file is now usable. PR by Philip Prindeville. GitHub #33.

2.2.2 (2016-01-21)
------------------

* `geoipupdate` now calls `fsync` on the database directory after a `rename`
  to make it durable in the event of a crash.

2.2.1 (2015-02-25)
------------------

* Bump version number to correct PPA release issue. No other changes to the
  source distribution.

2.2.0 (2015-02-25)
------------------

* `geoipupdate` now verifies the MD5 of the new database before deploying it.
  If the database MD5 does not match the expected MD5, `geoipupdate` will
  exit with an error.
* The copy of `base64.c` and `base64.h` was switched to a version under GPL 2+
  to prevent a license conflict.
* The `LICENSE` file was added to the distribution.
* Several issues in the documentation were fixed.

2.1.0 (2014-11-06)
------------------

* Previously `geoipupdate` did not check the status code of an HTTP response.
  It will now check for an unexpected status code and exit with a warning if
  such a status is received.
* The client now checks the return value of gz_close to ensure that the gzip
  stream was correctly decoded. GitHub PR #18.
* The client now checks that the file was correctly opened. Previous versions
  used an incorrect check.

2.0.2 (2014-07-22)
------------------

* The client now uses a single TCP connection when possible. Previously the
  public IP address of a host could change across requests, causing the
  authentication to fail. Reported by Aman Gupta. GitHub issue #12 and #13.
* `geoipupdate-pureperl.pl` was updated to work with GeoIP2.

2.0.1 (2014-05-02)
------------------

* Error handling was generally improved. `geoipupdate` will now return a 1
  whenever an update fails.
* Previously if one database failed to be updated, `geoipupdate` would not
  attempt to download the remaining databases. It now continues to the next
  database when a download fails.
* Support for Mac OS X 10.6, which is missing the `getline` function, was
  added.
* Unknown directives in the configuration file will now be logged.
* The debugging output was improved and made more readable.
* Several documentation errors and typos were fixed.

2.0.0 (2013-10-31)
------------------

* First stand-alone release.
