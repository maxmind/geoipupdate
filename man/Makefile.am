man_MANS = geoipupdate.1 GeoIP.conf.5

EXTRA_DIST = geoipupdate.1.in GeoIP.conf.5


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

UPDATE_MAN1 = $(mandir)/man1/geoipupdate.1

UPDATE_MAN5 = $(mandir)/man5/GeoIP.conf.5

install-data-hook:
	cat geoipupdate.1 | sed s,DATADIR,$(pkgdatadir), | sed s,CONF_DIR,$(sysconfdir), > $(DESTDIR)$(UPDATE_MAN1)
	cat GeoIP.conf.5 | sed s,DATADIR,$(pkgdatadir), | sed s,CONF_DIR,$(sysconfdir), > $(DESTDIR)$(UPDATE_MAN5)

