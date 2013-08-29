AM_CPPFLAGS = -I$(top_srcdir)/src -DSYSCONFDIR=\"$(sysconfdir)\" -Wall

#DEPS = $(top_builddir)/libGeoIP/libGeoIP.la
#LDADDS = $(top_builddir)/libGeoIP/libGeoIP.la

bin_PROGRAMS = geoipupdate

geoipupdate_SOURCES = geoipupdate.c
geoipupdate_LDFLAGS =
geoipupdate_DEPENDENCIES = $(top_builddir)/libGeoIP/libGeoIPUpdate.la
geoipupdate_LDADD = $(top_builddir)/src/libGeoIPUpdate.la
EXTRA_DIST = geoipupdate-pureperl.pl
