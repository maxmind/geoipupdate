# CHANGELOG

## 7.0.1 (2024-04-08)

* The 7.0.0 release was broken and has been retracted. This release updates the
  module path to match tagged version. Reported by Adam Weinberger. GitHub #306.

## 7.0.0 (2024-04-08)

* BREAKING CHANGE: Improvements to the HTTP download API.
  The client now calls two new endpoints:
  * `/geoip/updates/metadata` which is responsible for getting information about
    a database edition.
  * `/geoip/databases/{edition-id}/download` which is responsible for downloading
    the content of a database edition. This new endpoint redirects downloads to R2
    presigned URLs, so systems running `geoipupdate` need to be able to
    reach
    `mm-prod-geoip-databases.a2649acb697e2c09b632799562c076f2.r2.cloudflarestorage.com`
    in addition to `updates.maxmind.com`.
* BREAKING CHANGE: The public package API has been redesigned. The previous
  API was not easy to use and had become a maintenance burden. We now
  expose a `Client` at `github.com/maxmind/geoipupdate/client` with a
  `Download()` method. The intention is to expose less of the `geoipupdate`
  internals and provide a simpler and easier to use package. Many
  previously exposed methods and types are now either internal only or have
  been removed.
* BREAKING CHANGE: If set, `GEOIPUPDATE_VERBOSE` must either be `0` or `1`.
  All other values will return an error.
* Setting `GEOIPUPDATE_VERBOSE` to `1` now works as expected. In the 6.0.0 and
  6.1.0 releases, the flag was ignored. Reported by pmcevoy. GitHub #298.
* `geoipupdate` now supports retrying on more types of errors
  such as HTTP2 INTERNAL_ERROR.
* Now `geoipupdate` doesn't requires the user to specify the config file
  even if all the other arguments are set via the environment variables.
  Reported by jsf84ksnf. GitHub #284.

## 6.1.0 (2024-01-09)

* `geoipupdate` now sets the version in the `User-Agent` header to the
  version in the binary. While there were no issues with the version in the
  header, this makes sure it will match the binary. The header also now
  includes build information, such as OS and architecture.
* White spaces in secret files `GEOIPUPDATE_ACCOUNT_ID_FILE`,
  `GEOIPUPDATE_LICENSE_KEY_FILE` are ignored, see #262 issue.

## 6.0.0 (2023-07-12)

* `geoipupdate` now supports configuration via environment variables. Any
  configuration set this way will override any value from the config file,
  but still be overridden by any associated command line option (if any).
  The following new environment variables are supported:

    * `GEOIPUPDATE_ACCOUNT_ID`
    * `GEOIPUPDATE_ACCOUNT_ID_FILE`
    * `GEOIPUPDATE_CONF_FILE`
    * `GEOIPUPDATE_DB_DIR`
    * `GEOIPUPDATE_EDITION_IDS`
    * `GEOIPUPDATE_HOST`
    * `GEOIPUPDATE_LICENSE_KEY`
    * `GEOIPUPDATE_LICENSE_KEY_FILE`
    * `GEOIPUPDATE_LOCK_FILE`
    * `GEOIPUPDATE_PARALLELISM`
    * `GEOIPUPDATE_PRESERVE_FILE_TIMES`
    * `GEOIPUPDATE_PROXY`
    * `GEOIPUPDATE_PROXY_USER_PASSWORD`
    * `GEOIPUPDATE_RETRY_FOR`
    * `GEOIPUPDATE_VERBOSE`

* Changed the signature of `NewConfig` in `pkg/geoipupdate` to no longer accept
  a positional config file path argument, which can now be passed in using the
  option from `WithConfigFile` along with the other optional parameters.
* `geoipupdate` and `NewConfig` no longer require a config file to exist.
* The `--stack-trace` flag has been removed. This flag has been broken since
  4.11.0.

## 5.1.1 (2023-05-08)

* Based on feedback, the change to use a non-root user in 5.1.0
  when using the Docker image has been reverted. There are no
  non-Docker changes in this release. See GitHub #233.

## 5.1.0 (2023-05-05)

* Fixed the Docker health-check script to use the correct time of
  the last update attempt. Reported by cford1080. GitHub #225.
* Added new `--output` flag to print JSON to standard output describing
  the result of the run.
* Compilation with Go versions before 1.19 is no longer supported.
* When using the provided Docker images, `geoipupdate` no longer runs
  as root in the container. Based on pull request by Andreas
  Grünenfelder. GitHub #200.

## 5.0.4 (2023-04-17)

* On releases 4.9.0 through 5.0.3, the incorrect commit was tagged.
  This release attempts to fix the release process to prevent this
  issue. There are no code changes to the binaries provided by
  MaxMind, either on the GitHub Release page or the MaxMind PPA.

## 5.0.3 (2023-04-15)

* On 5.0.0 through 5.0.2, the default database directory was not being
  correctly set for Debian and RPM package builds. The directory
  `/usr/local/share/GeoIP` was being used rather than `/usr/share/GeoIP`.
  This build restores `/usr/share/GeoIP` as the default directory for
  these builds. Reported by Yvan. GitHub #222.

## 5.0.2 (2023-04-13)

* "Database ... up to date" messages are now only shown if the verbose
  flag is set. Pull request by Adam Weinberger. GitHub #219.

## 5.0.1 (2023-04-13)

* The 5.0.0 release mistakenly set the file permissions on downloaded
  databases to 0600. This restores the previous behavior of using 0644.
  Pull request by Josh Samuelson. GitHub #217 and #218.

## 5.0.0 (2023-04-12)

* Redefined the `Reader` and `Writer` interface APIs in
  `pkg/geoipupdate/database`. This change aims to to make it easier to
  introduce custom implementations of these interfaces.
* Changed the signature of `NewConfig` in `pkg/geoipupdate` to accept
  optional parameters. This change allows the introduction of new
  flags or config options without making breaking changes to the function's
  signature.
* Introduced `Parallelism` as a new flag and config option to enable
  concurrent database updates.

## 4.11.1 (2023-03-16)

* Removed extra underscore in script variables preventing the Docker
  secret support added in 4.11.0 from working as expected. Pull request by
  Moeen Mirjalili. GitHub #210.

## 4.11.0 (2023-03-15)

* `github.com/pkg/errors` is no longer used to wrap errors.
* Docker secrets are now supported for the MaxMind account ID and
  license key. Pull request by Matthew Kobayashi. GitHub #197.
* The Dockerfile now has a Healthcheck that makes sure the modification date
  of the database directory is within the update period.
* The Docker images are now published to the GitHub Container Registry,
  `ghcr.io`. We will likely stop publishing to Docker Hub in the near future.

## 4.10.0 (2022-09-26)

* HTTPS proxies are now supported. Pull request by Jamie Thompson. GitHub
  #172.
* An HTTP request to get the filename for the edition ID has been removed.
  This was previously required as the GeoIP Legacy edition IDs bore little
  relation to the name of the database on disk.

## 4.9.0 (2022-02-15)

* The client now sets the `User-Agent` header.
* The error handling has been improved.
* The `goreleaser` configuration has been consolidated. There is now
  one checksum file for all builds.
* Binaries are now built for OpenBSD and FreeBSD. Pull request by
  Devin Buhl. GitHub #161.
* Packages for ARM are now correctly uploaded. Bug report by Service Entity.
  GitHub #162.

## 4.8.0 (2021-07-20)

* The Docker container now supports the following new environment
  variables:

  * `GEOIPUPDATE_CONF_FILE` - The path where the configuration file will
    be written. The default is `/etc/GeoIP.conf`.
  * `GEOIPUPDATE_DB_DIR` - The directory where geoipupdate will download
    the databases. The default is `/usr/share/GeoIP`.

  Pull request by Maxence POULAIN. GitHub #143.

## 4.7.1 (2021-04-19)

* The Alpine version used for the Docker image now tracks the `alpine:3`
  tag rather than a specific point release.
* The `arm64` Docker images were not correctly generated in 4.7.0. This
  release corrects the issue.
* This release provides an `arm/v6` Docker image.

## 4.7.0 (2021-04-16)

* Go 1.13 or greater is now required.
* In verbose mode, we now print a message before each HTTP request.
  Previously we would not print anything for retried requests.
* Expected response errors no longer cause request retries. For example, we
  no longer retry the download request if the database subscription has
  lapsed.
* When running with `GEOIPUPDATE_FREQUENCY` set, the Docker image will now
  stop when sent a SIGTERM instead of waiting for a SIGKILL. Pull request
  by Maxence POULAIN. GitHub #135.
* Docker images are now provided for ARM64. Requested by allthesebugsv2.
  GitHub #136.

## 4.6.0 (2020-12-14)

* Show version number in verbose output.
* Retry downloads in more scenarios. Previously we would not retry failures
  occurring when reading the response body, but now we do.

## 4.5.0 (2020-10-28)

* We no longer use a third party library for exponential backoff. This
  restores support for older Go versions.

## 4.4.0 (2020-10-28)

* The edition ID is now included when there is a failure retrieving a
  database.
* The Docker image no longer prints the generated `GeoIP.conf` when starting
  up. This prevents a possible leak of the account's license key. Pull
  request by Nate Gay. GitHub #109.
* The minimum Go version is now 1.11.
* Failing HTTP requests are now retried using an exponential backoff. The
  period to keep retrying any failed request is set to 5 minutes by default and
  can be adjusted using the new `RetryFor` configuration option.
* When using the go package rather than the command-line tool, the null value
  for `RetryFor` will be 0 seconds, which means no retries will be performed. To
  change that, set `RetryFor` explicitly in the `Config` you provide, or obtain
  your `Config` value via `geoipupdate.NewConfig`.

## 4.3.0 (2020-04-16)

* First release to Docker Hub. Requested by Shun Yanaura. GitHub #24.
* The binary builds are now built with `CGO_ENABLED=0`. Request by CrazyMax.
  GitHub #63.

## 4.2.2 (2020-02-21)

* Re-release for PPA. No other changes.

## 4.2.1 (2020-02-21)

* The minimum Go version is now 1.10 again as this was needed to build the PPA
  packages.

## 4.2.0 (2020-02-20)

* The major version of the module is now included at the end of the module
  path. Previously, it was not possible to import the module in projects that
  were using Go modules. Reported by Roman Glushko. GitHub #81.
* The minimum Go version is now 1.13.
* A valid account ID and license key combination is now required for database
  downloads, so those configuration options are now required.
* The error handling when closing a local database file would previously
  ignore errors and, upon upgrading to `github.com/pkg/errors` 0.9.0,
  would fail to ignore expected errors. Reported by Ilya Skrypitsa and
  pgnd. GitHub #69 and #70.
* The RPM release was previously lacking the correct owner and group on files
  and directories. Among other things, this caused the package to conflict with
  the `GeoIP` package in CentOS 7 and `GeoIP-GeoLite-data` in CentOS 8. The
  files are now owned by `root`. Reported by neonknight. GitHub #76.

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
