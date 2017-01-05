DEFAULT_CONFIG_FILE = $(sysconfdir)/GeoIP.conf

EXTRA_DIST=GeoIP.conf.default

install-exec-hook:
	@if [ ! -d "$(DESTDIR)$(sysconfdir)" ]; then \
		echo "$(MKDIR_P) $(DESTDIR)$(sysconfdir)" ; \
		$(MKDIR_P) "$(DESTDIR)$(sysconfdir)" ; \
	fi
	@if [ -f "$(DESTDIR)$(DEFAULT_CONFIG_FILE)" ]; then \
	        echo "$@ will not overwrite existing $(DESTDIR)$(DEFAULT_CONFIG_FILE)" ; \
	else \
	        echo "$(INSTALL_DATA) GeoIP.conf.default $(DESTDIR)$(DEFAULT_CONFIG_FILE)"; \
	        $(INSTALL_DATA) "$(srcdir)/GeoIP.conf.default" "$(DESTDIR)$(DEFAULT_CONFIG_FILE)"; \
	fi

uninstall-hook:
	@if test -f "$(DESTDIR)$(DEFAULT_CONFIG_FILE)" ; then \
		rm "$(DESTDIR)$(DEFAULT_CONFIG_FILE)"; \
fi
