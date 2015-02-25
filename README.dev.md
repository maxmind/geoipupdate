Current PPA process:

1. Make an upstream/2.x.y branch off of the v2.x.y tag.
2. git co ubuntu-ppa
3. git merge v2.x.y
4. dch -i
  * Add new entry for utopic. Follow existing PPA versioning style.
5. Ensure your checkout is really clean. This includes ignored files. Our debian
   build does not currently ignore them.
6. gbp buildpackage -S
7. dput ppa:maxmind/ppa ../geoipupdate_<TAG>-<DEB VERSION>_source.changes

If 7 was successful, modify debian/changelog and repeate 6 & 7 for trusty and
precise.

