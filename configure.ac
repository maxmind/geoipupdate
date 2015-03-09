dnl AM_CONFIG_HEADER(config.h)

AC_INIT([geoipupdate], [2.2.2],[support@maxmind.com],[geoipupdate])
AC_GNU_SOURCE
AM_INIT_AUTOMAKE([foreign])
AC_CONFIG_SRCDIR([bin/geoipupdate.c])
GEOIPUPDATE_VERSION_INFO=`echo $VERSION | awk -F. '{ printf "%d:%d:%d", $1+$2, $3, $2 }'`
AC_SUBST(GEOIPUPDATE_VERSION_INFO)
AC_PREREQ

AC_PROG_CC_C99
AC_PROG_LIBTOOL
AC_PROG_MKDIR_P

AC_CHECK_TYPE(byte,[AC_DEFINE(HAVE_BYTE_TYPEDEF)],[])
AC_CHECK_TYPE(ushort,[AC_DEFINE(HAVE_USHORT_TYPEDEF)],[])
AC_CHECK_TYPE(ulong,[AC_DEFINE(HAVE_ULONG_TYPEDEF)],[])
AC_CHECK_TYPE(u16,[AC_DEFINE(HAVE_U16_TYPEDEF)],[])
AC_CHECK_TYPE(u32,[AC_DEFINE(HAVE_U32_TYPEDEF)],[])

AC_TYPE_SIZE_T
AC_TYPE_SSIZE_T
AC_TYPE_UINT64_T

AC_C_BIGENDIAN([AC_DEFINE(BIG_ENDIAN_HOST,1)],[AC_DEFINE(LITTLE_ENDIAN_HOST,1)])

AC_CHECK_HEADERS([stdint.h limits.h stddef.h])
AC_CHECK_HEADERS(zlib.h, , AC_MSG_ERROR(Zlib header (zlib.h) not found. You may need to install a zlib development package.))
AC_CHECK_HEADERS(curl/curl.h, , AC_MSG_ERROR(curl header (curl/curl.h) not found. You may need to install a curl development package.))

AC_CHECK_LIB(z, gzopen, [], [AC_MSG_ERROR([libz not found.])])
AC_CHECK_LIB(curl, curl_easy_setopt, [], [AC_MSG_ERROR([libcurl not found.])])

AC_CHECK_FUNCS([ getline fgets asprintf strcasecmp memset socket strchr strdup strstr strrchr memchr strspn])

AC_HEADER_STDBOOL

AC_C_INLINE
AC_C_RESTRICT

AC_OUTPUT([
Makefile
bin/Makefile
conf/Makefile
man/Makefile
])
