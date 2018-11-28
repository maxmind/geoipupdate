# Releasing

* Make sure you have [`hub`](https://github.com/github/hub) and
  [`goreleaser`](https://goreleaser.com/) installed.
* Update `CHANGELOG.md`. Set the appropriate release date.
* Run `GITHUB_TOKEN=<your token> ./dev-bin/release.sh`. For `goreleaser` you
  will need a token with the `repo` scope. You may create a token
  [here](https://github.com/settings/tokens/new).
