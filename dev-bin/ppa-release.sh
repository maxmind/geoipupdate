#!/bin/bash

set -e
set -x
set -u

DISTS=( lunar kinetic jammy focal bionic )

VERSION=$(perl -MFile::Slurper=read_text -MDateTime <<EOF
use v5.16;
my \$log = read_text(q{CHANGELOG.md});
\$log =~ /\n## (\d+\.\d+\.\d+) \((\d{4}-\d{2}-\d{2})\)\n/;
die 'Release time is not today!' unless DateTime->now->ymd eq \$2;
say \$1;
EOF
)

SRCDIST="geoipupdate-$VERSION.tar.gz"
SRC=/tmp/geoipupdate-$VERSION/
ORIG_NAME="geoipupdate_$VERSION.orig.tar.gz"
RESULTS=/tmp/build-geoipupdate-results/

rm -rf "$RESULTS" cmd/geoipupdate/geoipupdate build

make clean

go mod vendor

cp -a . "$SRC"
rm -f "$SRC"/*gz
tar --exclude=.git --exclude='*.swp' --exclude='*.gz' -C /tmp -czvf "$SRCDIST" "geoipupdate-$VERSION"

mkdir -p $RESULTS

for dist in "${DISTS[@]}"; do
    distdir=$(mktemp -d)
    cp -r "$SRC/" "$distdir/"
    cp "$SRCDIST" "$distdir/$ORIG_NAME"
    pushd "$distdir/geoipupdate-$VERSION/"
    dch -v "$VERSION-0+maxmind1~$dist" -D "$dist" -u low "New upstream release."
    # If you don't want to include the orig source, such as if you're just
    # bumping the PPA version (e.g. maxmind1 to maxmind2), replace the -sa flag
    # here with -sd. Note you will have to download the orig.tar.gz from
    # Launchpad to do this or it will reject the package saying it's different.
    # You can find the orig.tar.gz on the list of packages page on Launchpad.
    debuild -S -sa -rfakeroot
    popd
    ls "$distdir"
    mkdir -p "$RESULTS/$dist"
    cp "$distdir"/geoipupdate_* "$RESULTS/$dist/"
    cp "$distdir/geoipupdate-$VERSION/debian/changelog" "$RESULTS/$dist/changelog"
    rm -rf "$distdir"
done

read -e -p "Release to PPA? (y/n)" SHOULD_RELEASE

if [ "$SHOULD_RELEASE" != "y" ]; then
    echo "Aborting"
    exit 1
fi

dput ppa:maxmind/ppa "$RESULTS"/*/*source.changes


dch -v "$VERSION-0+maxmind1" -D "${DISTS[0]}" -u low "New upstream release."
git add debian/changelog
git commit -m "Update debian/changelog for $VERSION"
git push
