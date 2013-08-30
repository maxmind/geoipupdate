man_MANS = geoipupdate.1 geoiplookup6.1 geoiplookup.1

EXTRA_DIST = geoiplookup6.1.in geoiplookup.1.in geoipupdate.1.in


edit = sed \
	-e 's|DATADIR|$(pkgdatadir)|g' \
	-e 's|CONF_DIR|$(sysconfdir)|g'

geoipupdate.1 geoiplookup.1 geoiplookup6.1: Makefile
	rm -f $@ $@.tmp
	$(edit) '$(srcdir)/$@.in' >$@.tmp
	mv $@.tmp $@

geoipupdate.1: geoipupdate.1.in
geoiplookup.1: geoiplookup.1.in
geoiplookup6.1: geoiplookup6.1.in

CLEANFILES = geoiplookup6.1 geoipupdate.1 geoiplookup.1

UPDATE_MAN = $(mandir)/man1/geoipupdate.1
LOOKUP_MAN = $(mandir)/man1/geoiplookup.1
LOOKUP6_MAN = $(mandir)/man1/geoiplookup6.1

install-data-hook:
	cat geoipupdate.1 | sed s,DATADIR,$(pkgdatadir), | sed s,CONF_DIR,$(sysconfdir), > $(DESTDIR)$(UPDATE_MAN)
	cat geoiplookup.1 | sed s,DATADIR,$(pkgdatadir), > $(DESTDIR)$(LOOKUP_MAN)
	cat geoiplookup6.1 | sed s,DATADIR,$(pkgdatadir), > $(DESTDIR)$(LOOKUP6_MAN)

