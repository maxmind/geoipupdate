/* -*- Mode: C; indent-tabs-mode: t; c-basic-offset: 2; tab-width: 2 -*- */
/* GeoIP.h
 *
 * Copyright (C) 2006 MaxMind LLC
 *
 * This library is free software; you can redistribute it and/or
 * modify it under the terms of the GNU Lesser General Public
 * License as published by the Free Software Foundation; either
 * version 2.1 of the License, or (at your option) any later version.
 *
 * This library is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
 * Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public
 * License along with this library; if not, write to the Free Software
 * Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301  USA
 */

#ifndef GEOIPUPDATE_H
#define GEOIPUPDATE_H

#ifdef __cplusplus
extern "C" {
#endif

#include <stdarg.h>
#include <stdio.h>

typedef enum {
	GEOIP_NO_NEW_UPDATES          = 1,   /* Database up-to-date, no action taken */
  GEOIP_SUCCESS                 = 0,   /* Success */
  GEOIP_LICENSE_KEY_INVALID_ERR =	-1,  /* License Key Invalid */
  GEOIP_DNS_ERR                 = -11, /* Unable to resolve hostname */
  GEOIP_NON_IPV4_ERR            = -12, /* Non - IPv4 address */
  GEOIP_SOCKET_OPEN_ERR         = -13, /* Error opening socket */
	GEOIP_CONNECTION_ERR          = -14, /* Unable to connect */
	GEOIP_GZIP_IO_ERR             = -15, /* Unable to write GeoIP.dat.gz file */
  GEOIP_TEST_IO_ERR             = -16, /* Unable to write GeoIP.dat.test file */
	GEOIP_GZIP_READ_ERR           = -17, /* Unable to read gzip data */
	GEOIP_OUT_OF_MEMORY_ERR       = -18, /* Out of memory error */
	GEOIP_SOCKET_READ_ERR         = -19, /* Error reading from socket, see errno */
	GEOIP_SANITY_OPEN_ERR         = -20, /* Sanity check GeoIP_open error */
	GEOIP_SANITY_INFO_FAIL        = -21, /* Sanity check database_info string failed */
	GEOIP_SANITY_LOOKUP_FAIL      = -22, /* Sanity check ip address lookup failed */
	GEOIP_RENAME_ERR              = -23, /* Rename error while installing db, check errno */
	GEOIP_USER_ID_INVALID_ERR     = -24, /* Invalid userID */
	GEOIP_PRODUCT_ID_INVALID_ERR  = -25, /* Invalid product ID or subscription expired */
	GEOIP_INVALID_SERVER_RESPONSE = -26  /* Server returned invalid response */
} GeoIPUpdateCode;

const char * GeoIP_get_error_message(int i);

/* Original Update Function, just for MaxMind GeoIP Country database */
	short int GeoIP_update_database (char * license_key, int verbose, void (*f)( char *));

/* More generalized update function that works more databases */
	short int GeoIP_update_database_general (char * user_id, char * license_key,char * data_base_type, int verbose,char ** client_ipaddr, void (*f)( char *));

	/* experimental export */
	int  GeoIP_fprintf(int (*f)(FILE *, char *),FILE *fp, const char *fmt, ...);
	void GeoIP_printf(void (*f)(char *), const char *fmt, ...);

#ifdef __cplusplus
}
#endif

#endif /* GEOIPUPDATE_H */
