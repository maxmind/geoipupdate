# Releasing

Releases are built and published by
[`.github/workflows/release.yml`](.github/workflows/release.yml). Pushing a
`vX.Y.Z` tag is what releases it. Nothing is built or published from a developer
machine, so you do not need `goreleaser`, Docker, or a GitHub token locally.

- Set the release date in `CHANGELOG.md` and commit it to a release branch.
- Run `./dev-bin/release.sh`. It validates the tree, bumps
  `internal/vars/version.go`, commits, and pushes the tag.
- Watch the run: `gh run watch --exit-status`. It builds the archives, the deb
  and rpm packages, and the Docker images, publishes the GitHub release with the
  `CHANGELOG.md` entry as its body, pushes the images to Docker Hub and
  `ghcr.io`, and attaches build provenance attestations.
- If the run fails, fix the problem and re-run it from the Actions tab. The
  workflow's checks are all derived from the tagged tree, so a re-run does not
  go stale. If the tag itself was wrong, delete it locally and on `origin` and
  start over.

To try a release locally without publishing anything, install `goreleaser`
2.17.1 — the version
[`.github/workflows/release.yml`](.github/workflows/release.yml) pins, so use
the same one — and `pandoc`, then run
`goreleaser release --snapshot --clean --skip=publish,docker`. Skipping `docker`
avoids GoReleaser's Buildx invocations, which Podman's Docker-compatible CLI
does not fully support.

Then release to our PPA. This is still done by hand and is not part of the
workflow:

- Switch to the `ubuntu-ppa` branch. Merge the released tag into it. e.g.
  `git merge v4.1.0`.
- Set up to release to launchpad. You can see some information about
  prerequisites for this
  [here](https://github.com/maxmind/libmaxminddb/blob/main/README.dev.md).
- Ensure you have the `dh-golang`, `golang-any`, `devscripts`,
  `libfile-slurp-tiny-perl`, and `libdatetime-perl` packages installed.
- Delete `dist` directory.
- Check whether you need to update the `$DISTS` variable in
  `dev-bin/ppa-release.sh`. We should include all currently supported Ubuntu
  releases.
- Run `dev-bin/ppa-release.sh`

Gotcha with PPA:

- If you get an error from `dput` like
  `No host ppa:maxmind/ppa found in config`, you can create a `~/.dput.cf` with
  content like so:

```
[maxmind]
fqdn = ppa.launchpad.net
method = ftp
incoming = ~maxmind/ubuntu/ppa/
login = anonymous
allow_unsigned_uploads = 0
```

Then you can run the same `dput` command but with `dput maxmind [...]` instead
of `dput ppa:maxmind/ppa [...]` (I'm not sure how to make the matching work with
the original command).
