AM_CPPFLAGS = -I$(top_srcdir)/src -DSYSCONFDIR=\"$(sysconfdir)\" -DDATADIR=\"$(datadir)/GeoIP\" -Wall

bin_PROGRAMS = geoipupdate

geoipupdate_SOURCES = geoipupdate.c  md5.c base64.c geoipupdate_s.c product_s.c
geoipupdate_LDFLAGS =
geoipupdate_LDADD =
EXTRA_DIST = geoipupdate-pureperl.pl md5.h base64.h types.h geoiplookup.h
