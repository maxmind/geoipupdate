#!/bin/bash

set -eu -o pipefail

changelog=$(cat CHANGELOG.md)


if [[ -z ${GITHUB_TOKEN:-} ]]; then
    echo 'GITHUB_TOKEN must be set for goreleaser!'
    exit 1
fi

regex='
## ([0-9]+\.[0-9]+\.[0-9]+) \(([0-9]{4}-[0-9]{2}-[0-9]{2})\)

((.|
)*)'

if [[ ! $changelog =~ $regex ]]; then
      echo "Could not find date line in change log!"
      exit 1
fi

version="${BASH_REMATCH[1]}"
date="${BASH_REMATCH[2]}"
notes="$(echo "${BASH_REMATCH[3]}" | sed -n -e '/^## [0-9]\+\.[0-9]\+\.[0-9]\+/,$!p')"

if [[ "$date" !=  $(date +"%Y-%m-%d") ]]; then
    echo "$date is not today!"
    exit 1
fi

if [ -n "$(git status --porcelain)" ]; then
    echo ". is not clean." >&2
    exit 1
fi

tag="v$version"

echo $'\nRelease notes:'
echo "$notes"


read -p "Continue? (y/n) " ok

if [ "$ok" != "y" ]; then
    echo "Aborting"
    exit 1
fi

echo "Creating tag $tag"

message="$version

$notes"

git tag -a -m "$message" "$tag"

# goreleaser's `--rm-dist' should clear out `dist', but it didn't work for me.
rm -rf dist
goreleaser release --rm-dist -f .goreleaser.yml --release-notes <(echo "$message")
make clean BUILDDIR=.

rm -rf dist
goreleaser release --rm-dist -f .goreleaser-windows.yml --skip-publish
hub release edit -m "$message" \
    -a "dist/geoipupdate_${version}_windows_386.zip" \
    -a "dist/geoipupdate_${version}_windows_amd64.zip" \
    -a dist/checksums-windows.txt \
    "$tag"
make clean BUILDDIR=.

rm -rf dist
goreleaser release --rm-dist -f .goreleaser-packages.yml --skip-publish

git push
git push --tags

hub release edit -m "$message" \
    -a dist/checksums-dpkg-rpm.txt \
    -a "dist/geoipupdate_${version}_linux_386.deb" \
    -a "dist/geoipupdate_${version}_linux_amd64.deb" \
    -a "dist/geoipupdate_${version}_linux_386.rpm" \
    -a "dist/geoipupdate_${version}_linux_amd64.rpm" \
    "$tag"
make clean BUILDDIR=.
