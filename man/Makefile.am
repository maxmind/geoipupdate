man_MANS = geoipupdate.1

EXTRA_DIST = geoipupdate.1.in


edit = sed \
	-e 's|DATADIR|$(pkgdatadir)|g' \
	-e 's|CONF_DIR|$(sysconfdir)|g'

geoipupdate.1: Makefile
	rm -f $@ $@.tmp
	$(edit) '$(srcdir)/$@.in' >$@.tmp
	mv $@.tmp $@

geoipupdate.1: geoipupdate.1.in

CLEANFILES = geoipupdate.1

UPDATE_MAN = $(mandir)/man1/geoipupdate.1

install-data-hook:
	cat geoipupdate.1 | sed s,DATADIR,$(pkgdatadir), | sed s,CONF_DIR,$(sysconfdir), > $(DESTDIR)$(UPDATE_MAN)

