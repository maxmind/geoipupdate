dnl AM_CONFIG_HEADER(config.h)

AC_INIT([geoipupdate], [0.0.1],[support@maxmind.com],[geoipupdate])
AC_GNU_SOURCE
AM_INIT_AUTOMAKE
AC_CONFIG_SRCDIR([apps/geoipupdate.c])
GEOIP_VERSION_INFO=`echo $VERSION | awk -F. '{ printf "%d:%d:%d", $1+$2, $3, $2 }'`
AC_SUBST(GEOIP_VERSION_INFO)

AC_PREREQ

AC_PROG_CC_C99
AC_PROG_LIBTOOL

AC_CHECK_TYPE(byte,[AC_DEFINE(HAVE_BYTE_TYPEDEF)],[])
AC_CHECK_TYPE(ushort,[AC_DEFINE(HAVE_USHORT_TYPEDEF)],[])
AC_CHECK_TYPE(ulong,[AC_DEFINE(HAVE_ULONG_TYPEDEF)],[])
AC_CHECK_TYPE(u16,[AC_DEFINE(HAVE_U16_TYPEDEF)],[])
AC_CHECK_TYPE(u32,[AC_DEFINE(HAVE_U32_TYPEDEF)],[])

AC_TYPE_OFF_T
AC_TYPE_SIZE_T
AC_TYPE_SSIZE_T
AC_TYPE_UINT32_T
AC_TYPE_UINT64_T

AC_C_BIGENDIAN([AC_DEFINE(BIG_ENDIAN_HOST,1)],[AC_DEFINE(LITTLE_ENDIAN_HOST,1)])

#AC_CHECK_HEADERS([arpa/inet.h stdint.h netdb.h netinet/in.h sys/socket.h limits.h stddef.h])
AC_CHECK_HEADERS([stdint.h limits.h stddef.h])
AC_CHECK_HEADERS(zlib.h, , AC_MSG_ERROR(Zlib header (zlib.h) not found. You may need to install a zlib development package.))
AC_CHECK_HEADERS(curl/curl.h, , AC_MSG_ERROR(curl header (curl.h) not found. You may need to install a curl development package.))

#AC_CHECK_FUNC(gettimeofday, AC_DEFINE(HAVE_GETTIMEOFDAY))
#AC_CHECK_FUNC(vasprintf, AC_DEFINE(HAVE_VASPRINTF))
#AC_CHECK_FUNC(vsnprintf, AC_DEFINE(HAVE_VSNPRINTF))
#AC_CHECK_FUNC(vsprintf, AC_DEFINE(HAVE_VSPRINTF))

AC_CHECK_FUNCS([ asprintf strcasecmp memset socket strchr strdup strstr strrchr memchr strspn])

AC_CHECK_HEADER_STDBOOL

#AC_FUNC_MALLOC
#AC_FUNC_REALLOC
AC_C_INLINE
AC_C_RESTRICT

AC_OUTPUT([
Makefile
src/Makefile
apps/Makefile
man/Makefile
])
