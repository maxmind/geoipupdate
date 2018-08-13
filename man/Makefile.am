man_MANS = geoipupdate.1 GeoIP.conf.5

EXTRA_DIST = geoipupdate.1.in GeoIP.conf.5.in

edit = sed \
	-e 's|DATADIR|$(datadir)/GeoIP|g' \
	-e 's|CONF_DIR|$(sysconfdir)|g'

GeoIP.conf.5: Makefile
	rm -f $@ $@.tmp
	$(edit) '$(srcdir)/$@.in' >$@.tmp
	mv $@.tmp $@

GeoIP.conf.5: GeoIP.conf.5.in

geoipupdate.1: Makefile
	rm -f $@ $@.tmp
	$(edit) '$(srcdir)/$@.in' >$@.tmp
	mv $@.tmp $@

geoipupdate.1: geoipupdate.1.in

CLEANFILES = geoipupdate.1 GeoIP.conf.5
