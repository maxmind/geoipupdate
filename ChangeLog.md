GeoIP Update Change Log
=======================

2.0.2 (2014-07-22)
------------------

* The client now uses a single TCP connection when possible. Previously the
  public IP address of a host could change across requests, causing the
  authentication to fail. Reported by Aman Gupta. GitHub issue #12 and #13.
* ` geoipupdate-pureperl.pl` was updated to work with GeoIP2.

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
