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

if ! grep -q "^module github.com/maxmind/geoipupdate/$(echo "$tag" |cut -d . -f 1)" go.mod; then
    echo "Tag version does not match go.mod version!"
    exit 1;
fi

perl -pi -e "s/(?<=Version = \").+?(?=\")/$version/g" internal/vars/version.go

echo $'\nRelease notes:'
echo "$notes"

read -p "Continue? (y/n) " ok

if [ "$ok" != "y" ]; then
    echo "Aborting"
    exit 1
fi

if [ -n "$(git status --porcelain)" ]; then
    git commit -m "Update for $tag" -a
fi

git push

echo "Creating tag $tag"

message="$version

$notes"

git tag -a -m "$message" "$tag"

git push

# goreleaser's `--clean' should clear out `dist', but it didn't work for me.
rm -rf dist
goreleaser release --clean -f .goreleaser.yml --release-notes <(echo "$notes")
