#!/bin/bash

set -e

VERSION=$(perl -MFile::Slurp::Tiny=read_file -MDateTime <<EOF
use v5.16;
my \$log = read_file(q{ChangeLog.md});
\$log =~ /\n(\d+\.\d+\.\d+) \((\d{4}-\d{2}-\d{2})\)\n/;
die 'Release time is not today!' unless DateTime->now->ymd eq \$2;
say \$1;
EOF
)

TAG="v$VERSION"

git pull

if [ -n "$(git status --porcelain)" ]; then
    echo ". is not clean." >&2
    exit 1
fi

export VERSION
perl -pi -e "s/(?<=^AC_INIT\(\[geoipupdate\], \[).+?(?=\])/$VERSION/g" configure.ac

./bootstrap
./configure
make clean
make dist

git diff

read -e -p "Push to origin? " SHOULD_PUSH

if [ "$SHOULD_PUSH" != "y" ]; then
    echo "Aborting"
    exit 1
fi

git commit -m "Update version number for $TAG" -a

git tag -a "$TAG"
git push
git push --tags

echo "Release to PPA and Homebrew"
