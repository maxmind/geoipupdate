lib_LTLIBRARIES = libGeoIPUpdate.la

EXTRA_DIST = md5.h base64.h global.h types.h

AM_CPPFLAGS = -DGEOIPDATADIR=\"$(pkgdatadir)\" -Wall

include_HEADERS = GeoIPUpdate.h

libGeoIPUpdate_la_SOURCES = GeoIPUpdate.c md5.c base64.c

libGeoIP_la_LDFLAGS = -no-undefined -version-info @GEOIP_VERSION_INFO@

libGeoIPUpdate_la_LIBADD = -lz 

GeoIPUpdate.lo GeoIPUpdate.o: GeoIPUpdate.c

md5.lo md5.o: md5.c

base64.lo base64.o: base64.c

