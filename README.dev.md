Current PPA process:

1. Type `dch -i` and add the appropriate `debian/changelog` entry.
2. Move tarball created above to a temp directory and
   name it `geoipupdate_1.?.?.orig.tar.gz`.
3. Unpack tarball.
4. Copy `debian` directory from the ubuntu-ppa branch. We do not include
   it in the tarball or in master so that we don't interfere with the
   packaging done by Debian.
5. Update `debian/changelog` for the dist you are releasing to, e.g.,
   precise, trusty, utopic, and prefix the version with the a `~` followed
   by the dist name, e.g., `2.1.0-1+maxmind1~trusty`.
6. Run `debuild -S -sa -rfakeroot -k<KEY>`. (The key may not be necessary
   if your .bashrc is appropriately )
7. Check the lintian output to make sure everything looks sane.
8. Run `dput ppa:maxmind/ppa ../<source.changes files created above>` to
   upload.
9. Repeat 4-8 for remaining distributions.

This dist is _not_ yet buildable with gbp. You must build from the tarball.
