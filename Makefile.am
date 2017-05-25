AUTOMAKE_OPTIONS = foreign

SUBDIRS =   \
    bin     \
    conf    \
    man

EXTRA_DIST = ChangeLog.md LICENSE README.md bootstrap

doc_DATA = ChangeLog.md LICENSE README.md conf/GeoIP.conf.default

install-data-local:
	$(MKDIR_P) "$(DESTDIR)$(datarootdir)/GeoIP"
