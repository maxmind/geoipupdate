# Releasing

* Make sure you have [`hub`](https://github.com/github/hub),
  [`goreleaser`](https://goreleaser.com/), and rpmbuild installed.
  (rpmbuild is in the Ubuntu package `rpm`).
* Update `CHANGELOG.md`. Set the appropriate release date.
* Run `GITHUB_TOKEN=<your token> ./dev-bin/release.sh`. For `goreleaser` you
  will need a token with the `repo` scope. You may create a token
  [here](https://github.com/settings/tokens/new).
  * If we're not using Go modules yet, the release might fail depending on
    your `GO111MODULE` setting. Consider setting it to `off` if necessary.

Then release to our PPA:

* Switch to the ubuntu-ppa branch. Merge master into it.
* Set up to release to launchpad. You can see some information about
  prerequisites for this
  [here](https://github.com/maxmind/libmaxminddb/blob/master/README.dev.md).
* Run `dev-bin/ppa-release.sh`
