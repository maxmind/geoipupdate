# CHANGELOG

## 4.1.5 (2019-11-08)

* Respect the defaultConfigFile and defaultDatabaseDirectory variables in
  the main package again. They were ignored in 4.1.0 through 4.1.4. If not
  specified, the GitHub and PPA releases for these versions used the config
  /usr/local/etc/GeoIP.conf instead of /etc/GeoIP.conf and the database
  directory /usr/local/share/GeoIP instead of /usr/share/GeoIP.

## 4.1.4 (2019-11-07)

* Re-release of 4.1.3 as two commits were missing. No changes.

## 4.1.3 (2019-11-07)

* Remove formatting, linting, and testing from the geoipupdate target in
  the Makefile.

## 4.1.2 (2019-11-07)

* Re-release of 4.1.1 to fix Ubuntu PPA release issue. No code changes.

## 4.1.1 (2019-11-07)

* Re-release of 4.1.0 to fix Ubuntu PPA release issue. No code changes.

## 4.1.0 (2019-11-07)

* Improve man page formatting and organization. Pull request by Faidon
  Liambotis. GitHub #44.
* Provide update functionality as an importable package as well as a
  standalone program. Pull request by amzhughe. GitHub #48.

## 4.0.6 (2019-09-13)

* Re-release of 4.0.5 to fix Ubuntu PPA release issue. No code changes.

## 4.0.5 (2019-09-13)

* Ignore errors when syncing file system. These errors were primarily due
  to the file system not supporting the sync call. Reported by devkappa.
  GitHub #37.
* Use CRLF line endings on Windows for text files.
* Fix tests on Windows.
* Improve man page formatting. Reported by Faidon Liambotis. GitHub #38.
* Dependencies are no longer vendored. Reported by Faidon Liambotis. GitHub
  #39.

## 4.0.4 (2019-08-30)

* Do not try to sync the database directory when running on Windows.
  Syncing this way is not supported there and would lead to an error. Pull
  request by Nicholi. GitHub #32.

## 4.0.3 (2019-06-07)

* Update flock dependency from `theckman/go-flock` to `gofrs/flock`. Pull
  request by Paul Howarth. GitHub #22.
* Switch to Go modules and update dependencies.
* Fix version output on Ubuntu PPA and Homebrew releases.

## 4.0.2 (2019-01-18)

* Fix dependency in `Makefile`.

## 4.0.1 (2019-01-17)

* Improve documentation.
* Add script to generate man pages to `Makefile`.

## 4.0.0 (2019-01-14)

* Expand installation instructions.
* First full release.

## 0.0.2 (2018-11-28)

* Fix the output when the version output, `-V`, is passed to `geoipupdate`.

## 0.0.1 (2018-11-27)

* Initial version
