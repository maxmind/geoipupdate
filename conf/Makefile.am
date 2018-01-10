DEFAULT_CONFIG_FILE = $(sysconfdir)/GeoIP.conf

EXTRA_DIST=GeoIP.conf.default

edit = sed -e 's|DATADIR|$(datadir)/GeoIP|g'

install-exec-hook:
	@if [ ! -d "$(DESTDIR)$(sysconfdir)" ]; then \
		echo "$(MKDIR_P) $(DESTDIR)$(sysconfdir)" ; \
		$(MKDIR_P) "$(DESTDIR)$(sysconfdir)" ; \
	fi
	@if [ -f "$(DESTDIR)$(DEFAULT_CONFIG_FILE)" ]; then \
		echo "$@ will not overwrite existing $(DESTDIR)$(DEFAULT_CONFIG_FILE)" ; \
	else \
		echo "$(INSTALL_DATA) GeoIP.conf $(DESTDIR)$(DEFAULT_CONFIG_FILE)"; \
		$(edit) "$(srcdir)/GeoIP.conf.default" > "$(srcdir)/GeoIP.conf"; \
		$(INSTALL_DATA) "$(srcdir)/GeoIP.conf" "$(DESTDIR)$(DEFAULT_CONFIG_FILE)"; \
		rm "$(srcdir)/GeoIP.conf"; \
	fi

uninstall-hook:
	@if test -f "$(DESTDIR)$(DEFAULT_CONFIG_FILE)" ; then \
		rm "$(DESTDIR)$(DEFAULT_CONFIG_FILE)"; \
	fi
