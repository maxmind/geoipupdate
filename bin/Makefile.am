AM_CPPFLAGS = -I$(top_srcdir)/src -DSYSCONFDIR=\"$(sysconfdir)\" -DDATADIR=\"$(datadir)/GeoIP\" \
			  -std=c99 -Wall -Wextra

bin_PROGRAMS = geoipupdate

geoipupdate_SOURCES = \
					  geoipupdate.c geoipupdate.h \
					  md5.c md5.h \
					  base64.c base64.h \
					  geoipupdate_s.c \
					  product_s.c \
					  types.h \
					  functions.c functions.h
geoipupdate_LDFLAGS =
geoipupdate_LDADD =

EXTRA_DIST = geoipupdate-pureperl.pl

TESTS = functions_test

check_PROGRAMS = functions_test

functions_test_SOURCES = functions.c
functions_test_CPPFLAGS = $(AM_CPPFLAGS) -DTEST_FUNCTIONS
