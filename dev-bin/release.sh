#!/bin/bash

set -eu -o pipefail

changelog=$(cat CHANGELOG.md)

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

if [[ "$date" -ne  $(date +"%Y-%m-%d") ]]; then
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

git push
git push --tags

goreleaser release --rm-dist -f .goreleaser.yml --release-notes <(echo "$message")
goreleaser release --rm-dist -f .goreleaser-windows.yml --release-notes <(echo"$message")
goreleaser release --rm-dist -f .goreleaser-packages.yml --release-notes <(echo "$message")
