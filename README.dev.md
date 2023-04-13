# Releasing

* Make sure you have [`goreleaser`](https://goreleaser.com/), rpmbuild,
  and pandoc installed. (rpmbuild is in the Ubuntu package `rpm`).
* Set release date in `CHANGELOG.md` and commit it.
* Ensure you can run `docker` commands as your user (e.g., `docker
  images`).
* Log in to your Docker Hub account (be sure to be in our organization):
  `docker login`.
* Run `GITHUB_TOKEN=<your token> ./dev-bin/release.sh`. For `goreleaser` you
  will need a token with the `repo` scope. You may create a token
  [here](https://github.com/settings/tokens/new).

Then release to our PPA:

* Switch to the `ubuntu-ppa` branch. Merge the released tag into it. e.g.
  `git merge v4.1.0`.
* Set up to release to launchpad. You can see some information about
  prerequisites for this
  [here](https://github.com/maxmind/libmaxminddb/blob/main/README.dev.md).
* Ensure you have the `dh-golang`, `golang-any`, `devscripts`,
  `libfile-slurp-tiny-perl`, and `libdatetime-perl` packages installed.
* Delete `dist` directory.
* Check whether you need to update the `$DISTS` variable in
  `dev-bin/ppa-release.sh`. We should include all currently supported
  Ubuntu releases.
* Run `dev-bin/ppa-release.sh`
* Visit https://launchpad.net/~maxmind/+archive/ubuntu/ppa/+copy-packages to
  copy binaries to any versions that we support but don't build on, e.g.,
  Xenial.

Gotcha with PPA:

* If you get an error from `dput` like `No host ppa:maxmind/ppa found in
  config`, you can create a `~/.dput.cf` with content like so:

```
[maxmind]
fqdn = ppa.launchpad.net
method = ftp
incoming = ~maxmind/ubuntu/ppa/
login = anonymous
allow_unsigned_uploads = 0
```

Then you can run the same `dput` command but with `dput maxmind [...]`
instead of `dput ppa:maxmind/ppa [...]` (I'm not sure how to make the
matching work with the original command).
