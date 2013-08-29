/* -*- Mode: C; indent-tabs-mode: t; c-basic-offset: 2; tab-width: 2 -*- */
/* GeoIPUpdate.c
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

#include "base64.h"

#include "GeoIPCity.h"
#include "GeoIP.h"
#include "GeoIPUpdate.h"
#include "GeoIP_internal.h"

#include "global.h"
#include "md5.h"
#include <sys/types.h>
#if !defined(_WIN32)
#include <netinet/in.h>
#include <arpa/inet.h>
#include <sys/socket.h>
#include <netdb.h>
#else
#include <windows.h>
#include <winsock.h>
#endif
#include <zlib.h>
#include <time.h>
#include <stdio.h>
#include <unistd.h>

#ifdef _UNUSED
#elif defined(__GNUC__)
#define _UNUSED __attribute__ ((unused))
#else
#define _UNUSED
#endif

#define BLOCK_SIZE 1024

/* Update DB Host & HTTP GET Request formats:
 * ------------------------------------------
 * GET must support an optional HTTP Proxy.
 */
const char *GeoIPUpdateHost = "updates.maxmind.com";
/* This is the direct, or proxy port number. */
static int GeoIPHTTPPort = 80;
/* License-only format (OLD) */
const char *GeoIPHTTPRequest = "GET %s%s/app/update?license_key=%s&md5=%s HTTP/1.0\r\nHost: updates.maxmind.com\r\n";
/* General DB Types formats */
const char *GeoIPHTTPRequestFilename = "GET %s%s/app/update_getfilename?product_id=%s HTTP/1.0\r\nHost: %s\r\n";
const char *GeoIPHTTPRequestClientIP = "GET %s%s/app/update_getipaddr HTTP/1.0\r\nHost: %s\r\n";
const char *GeoIPHTTPRequestMD5 = "GET %s%s/app/update_secure?db_md5=%s&challenge_md5=%s&user_id=%s&edition_id=%s HTTP/1.0\r\nHost: updates.maxmind.com\r\n";
const char *ProxyAuthorization = "Proxy-Authorization: Basic %s\r\n";

/* messages */
const char *NoCurrentDB = "%s can't be opened, proceeding to download database\n";
const char *MD5Info = "MD5 Digest of installed database is %s\n";
const char *SavingGzip = "Saving gzip file to %s ... ";
const char *WritingFile = "Writing uncompressed data to %s ...";

const char * GeoIP_get_error_message(int i) {
  switch (i) {
  case GEOIP_NO_NEW_UPDATES:
    return "no new updates";
  case GEOIP_SUCCESS:
    return "Success";
  case GEOIP_LICENSE_KEY_INVALID_ERR:
    return "License Key Invalid";
  case GEOIP_DNS_ERR:
    return "Unable to resolve hostname";
  case GEOIP_NON_IPV4_ERR:
    return "Non - IPv4 address";
  case GEOIP_SOCKET_OPEN_ERR:
    return "Error opening socket";
  case GEOIP_CONNECTION_ERR:
    return "Unable to connect";
  case GEOIP_GZIP_IO_ERR:
    return "Unable to write GeoIP.dat.gz file";
  case GEOIP_TEST_IO_ERR:
    return "Unable to write GeoIP.dat.test file";
  case GEOIP_GZIP_READ_ERR:
    return "Unable to read gzip data";
  case GEOIP_OUT_OF_MEMORY_ERR:
    return "Out of memory error";
  case GEOIP_SOCKET_READ_ERR:
    return "Error reading from socket, see errno";
  case GEOIP_SANITY_OPEN_ERR:
    return "Sanity check GeoIP_open error";
  case GEOIP_SANITY_INFO_FAIL:
    return "Sanity check database_info string failed";
  case GEOIP_SANITY_LOOKUP_FAIL:
    return "Sanity check ip address lookup failed";
  case GEOIP_RENAME_ERR:
    return "Rename error while installing db, check errno";
  case GEOIP_USER_ID_INVALID_ERR:
    return "Invalid userID";
  case GEOIP_PRODUCT_ID_INVALID_ERR:
    return "Invalid product ID or subscription expired";
  case GEOIP_INVALID_SERVER_RESPONSE:
    return "Server returned something unexpected";
  default:
    return "no error";
  }  
}
int GeoIP_fprintf(int (*f)(FILE *, char *),FILE *fp, const char *str, ...) {
  va_list ap;
  int rc;
  char * f_str;
  int silence _UNUSED;

  if ( f == NULL )
    return 0;
  va_start(ap, str);
#if defined(HAVE_VASPRINTF)
  silence = vasprintf(&f_str, str, ap);
#elif defined (HAVE_VSNPRINTF)
  f_str = malloc(4096);
  if ( f_str )
    silence = vsnprintf(f_str, 4096, str, ap);
#else
  f_str = malloc(4096);
  if ( f_str )
    silence = vsprintf(f_str, str, ap);
#endif
  va_end(ap);
  if (  f_str == NULL )
    return -1;
  rc = (*f)(fp, f_str);
  free(f_str);
  return(rc);
}

void GeoIP_printf(void (*f)(char *), const char *str,...) {
  va_list params;
  char * f_str;
  int silence _UNUSED;
  if (f == NULL)
    return;
  va_start(params, str);
#if defined(HAVE_VASPRINTF)
  silence = vasprintf(&f_str, str, params);
#elif defined (HAVE_VSNPRINTF)
  f_str = malloc(4096);
  if ( f_str )
    silence = vsnprintf(f_str, 4096, str, params);
#else
  f_str = malloc(4096);
  if ( f_str )
    silence = vsprintf(f_str, str, params);
#endif
  va_end(params);
  if ( f_str == NULL )
    return;
  (*f)(f_str);
  free(f_str);
}

/* Support HTTP Proxy Host
 * --------------------------------------------------
 * Use typical OS support for the optional HTTP Proxy.
 *
 * Proxy adds http://{real-hostname} to URI format strings:
 * sprintf("GET %s%s/ HTTP/1.1\r\n",GeoIPProxyHTTP,GeoIPProxiedHost, ...);
 */

/* The Protocol is usually "" OR "http://" with a proxy. */
static char *GeoIPProxyHTTP = "";
/* GeoIP Hostname where proxy forwards requests. */
static char *GeoIPProxiedHost = "";

/* base64-encoded username and password that may be required by the proxy */
static char *GeoIPProxyCreds = NULL;

/* Read http_proxy env. variable & parse it.
 * -----------------------------------------
 * Allow only these formats:
 * "http://server.com", "http://server.com:8080"
 * OR
 * "server.com", "server.com:8080"
 *
 * A "user:password@" part will break this.
 */
short int parse_http_proxy(char **proxy_host, char **proxy_creds, int *port) {
	char * http_proxy;
	char * port_value;
	char * at_sign;

	if ((http_proxy = getenv("http_proxy"))) {

		if (! strncmp("http://", http_proxy, 7)) http_proxy += 7;

		*proxy_host = strdup(http_proxy);
		if ( *proxy_host == NULL )
			return 0; /* let the other functions deal with the memory error */

		if ((port_value = strrchr(*proxy_host, ':'))) {
			*port_value++ = '\0';
			*port = atoi(port_value);
		}
		else {
			*port = 80;
		}

		if ((at_sign = strchr(*proxy_host,'@'))) {
			*proxy_creds = *proxy_host;
			*proxy_host = at_sign +1;
			*at_sign = '\0';
		} else {
			*proxy_creds = NULL;
		}

		return(1);
	}
	else {
		return(0);
	}
}

/* Get the GeoIP host or the current HTTP Proxy host. */
struct hostent *GeoIP_get_host_or_proxy ( void (*f)( char * ) ) {
	char * hostname = (char *) GeoIPUpdateHost;
	char * proxy_host;
	char * proxy_creds;
	char * encoded_proxy_creds;
	size_t encoded_proxy_creds_len;
	int proxy_port;

	/* Set Proxy from OS: Unix/Linux */
	if (parse_http_proxy(&proxy_host, &proxy_creds, &proxy_port)) {

		GeoIPProxyHTTP = "http://";
		hostname = proxy_host;

             if ( proxy_creds == NULL )
                  proxy_creds = "";

		// The current code assumes there are no reserved/unsafe characters in the username or password.
		// The username and password should be URL decoding before they are base64-encoded for the Proxy-Authorization
		encoded_proxy_creds_len = base64_encode_alloc(proxy_creds, strlen(proxy_creds), &encoded_proxy_creds);
		if (encoded_proxy_creds == NULL) {
			if (encoded_proxy_creds_len == 0 && strlen(proxy_creds) != 0) {
				GeoIP_printf(f,"Error processing proxy credentials: data too long: %d", strlen(proxy_creds));
			} else {
				GeoIP_printf(f,"Error processing proxy credentials: out of memory");
			}
		} else {
			GeoIPProxyCreds = malloc(sizeof(char) * (strlen(ProxyAuthorization) + strlen(encoded_proxy_creds) + 1)); 
			sprintf(GeoIPProxyCreds, ProxyAuthorization, encoded_proxy_creds);
			GeoIPProxiedHost = (char *) GeoIPUpdateHost;
			GeoIPHTTPPort = proxy_port;
		}

		free(encoded_proxy_creds);

	}

	/* Resolve DNS host entry. */
	return(gethostbyname(hostname));
}

void GeoIP_send_request_uri(const int sock, const char *request_uri) {

	send(sock, request_uri, strlen(request_uri),0);
	if (GeoIPProxyCreds) {
		send(sock,GeoIPProxyCreds,strlen(GeoIPProxyCreds),0);
	}
	send(sock, "\r\n", 2 ,0);
}

short int GeoIP_update_database (char * license_key, int verbose, void (*f)( char * )) {
	struct hostent *hostlist;
	int sock;
	char * buf, *tmp;
	struct sockaddr_in sa;
	int offset = 0, err;
	char * request_uri;
	char * compr;
	unsigned long comprLen;
	FILE *comp_fh, *cur_db_fh, *gi_fh;
	gzFile gz_fh;
	char * file_path_gz, * file_path_test;
	MD5_CONTEXT context;
	unsigned char buffer[1024], digest[16];
	char hex_digest[33] = "00000000000000000000000000000000\0";
	unsigned int i;
	GeoIP * gi;
	char * db_info;
	char block[BLOCK_SIZE];
	int block_size = BLOCK_SIZE;
	size_t len;
	size_t written;
	_GeoIP_setup_dbfilename();

	/* get MD5 of current GeoIP database file */
	if ((cur_db_fh = fopen (GeoIPDBFileName[GEOIP_COUNTRY_EDITION], "rb")) == NULL) {
    GeoIP_printf(f, NoCurrentDB, GeoIPDBFileName[GEOIP_COUNTRY_EDITION]);
	} else {
		md5_init(&context);
		while ((len = fread (buffer, 1, 1024, cur_db_fh)) > 0)
			md5_write (&context, buffer, len);
		md5_final (&context);
		memcpy(digest,context.buf,16);
		fclose (cur_db_fh);
		for (i = 0; i < 16; i++) {
			// "%02x" will write 3 chars
			snprintf (&hex_digest[2*i], 3, "%02x", digest[i]);
		}
    GeoIP_printf(f, MD5Info, hex_digest);
	}

	hostlist = GeoIP_get_host_or_proxy(f);

	if (hostlist == NULL)
		return GEOIP_DNS_ERR;

	if (hostlist->h_addrtype != AF_INET)
		return GEOIP_NON_IPV4_ERR;

	if((sock = socket(AF_INET, SOCK_STREAM, 0)) < 0) {
		return GEOIP_SOCKET_OPEN_ERR;
	}

	memset(&sa, 0, sizeof(struct sockaddr_in));
	sa.sin_port = htons(GeoIPHTTPPort);
	memcpy(&sa.sin_addr, hostlist->h_addr_list[0], hostlist->h_length);
	sa.sin_family = AF_INET;

	if (verbose == 1){
		GeoIP_printf(f,"Connecting to MaxMind GeoIP Update server\n");
		GeoIP_printf(f, "via Host or Proxy Server: %s:%d\n", hostlist->h_name, GeoIPHTTPPort);
	}	

	/* Download gzip file */
	if (connect(sock, (struct sockaddr *)&sa, sizeof(struct sockaddr))< 0)
		return GEOIP_CONNECTION_ERR;

	request_uri = malloc(sizeof(char) * (strlen(license_key) + strlen(GeoIPHTTPRequest)
                              + strlen(GeoIPProxyHTTP) + strlen(GeoIPProxiedHost) + 36 + 1));
	if (request_uri == NULL)
		return GEOIP_OUT_OF_MEMORY_ERR;
	sprintf(request_uri,GeoIPHTTPRequest,GeoIPProxyHTTP,GeoIPProxiedHost,license_key, hex_digest);

	GeoIP_send_request_uri(sock, request_uri);
	free(request_uri);

	buf = malloc(sizeof(char) * block_size + 1);
	if (buf == NULL)
		return GEOIP_OUT_OF_MEMORY_ERR;

	if (verbose == 1)
		GeoIP_printf(f,"Downloading gzipped GeoIP Database...\n");

	for (;;) {
		int amt;
		amt = recv(sock, &buf[offset], block_size,0);
		if (amt == 0) {
			break;
		} else if (amt == -1) {
			free(buf);
			return GEOIP_SOCKET_READ_ERR;
		}
		offset += amt;
		tmp = buf;
		buf = realloc(buf, offset+block_size + 1);
		if (buf == NULL){
		        free(tmp);
			return GEOIP_OUT_OF_MEMORY_ERR;
		}
	}

        buf[offset]=0;
	compr = strstr(buf, "\r\n\r\n");
        if ( compr == NULL ) {
   		free(buf);
		return GEOIP_INVALID_SERVER_RESPONSE;
        }
        /* skip searchstr  "\r\n\r\n" */
        compr += 4;
	comprLen = offset + buf - compr;

	if (strstr(compr, "License Key Invalid") != NULL) {
		if (verbose == 1)
			GeoIP_printf(f,"Failed\n");
		free(buf);
		return GEOIP_LICENSE_KEY_INVALID_ERR;
	} else if (strstr(compr, "Invalid product ID or subscription expired") != NULL){
		free(buf);
		return GEOIP_PRODUCT_ID_INVALID_ERR;
	} else if (strstr(compr, "No new updates available") != NULL) {
		free(buf);
		return GEOIP_NO_NEW_UPDATES;
	}

	if (verbose == 1)
		GeoIP_printf(f,"Done\n");

	/* save gzip file */
	file_path_gz = malloc(sizeof(char) * (strlen(GeoIPDBFileName[GEOIP_COUNTRY_EDITION]) + 4));
	if (file_path_gz == NULL)
		return GEOIP_OUT_OF_MEMORY_ERR;
	strcpy(file_path_gz,GeoIPDBFileName[GEOIP_COUNTRY_EDITION]);
	strcat(file_path_gz,".gz");
	if (verbose == 1) {
    GeoIP_printf(f, SavingGzip, file_path_gz);
	}
	comp_fh = fopen(file_path_gz, "wb");

	if(comp_fh == NULL) {
		free(file_path_gz);
		free(buf);
		return GEOIP_GZIP_IO_ERR;
	}

	written = fwrite(compr, 1, comprLen, comp_fh);
	fclose(comp_fh);
	free(buf);

        if ( written != comprLen )
		return GEOIP_GZIP_IO_ERR;

	if (verbose == 1)
		GeoIP_printf(f,"Done\n");

	if (verbose == 1)
		GeoIP_printf(f,"Uncompressing gzip file ... ");

	/* uncompress gzip file */
	gz_fh = gzopen(file_path_gz, "rb");
	file_path_test = malloc(sizeof(char) * (strlen(GeoIPDBFileName[GEOIP_COUNTRY_EDITION]) + 6));
	if (file_path_test == NULL)
		return GEOIP_OUT_OF_MEMORY_ERR;
	strcpy(file_path_test,GeoIPDBFileName[GEOIP_COUNTRY_EDITION]);
	strcat(file_path_test,".test");
	gi_fh = fopen(file_path_test, "wb");

	if(gi_fh == NULL) {
		free(file_path_test);
		return GEOIP_TEST_IO_ERR;
	}
	for (;;) {
		int amt;
		amt = gzread(gz_fh, block, block_size);
		if (amt == -1) {
			free(file_path_test);
			fclose(gi_fh);
			gzclose(gz_fh);
			return GEOIP_GZIP_READ_ERR;
		}
		if (amt == 0) {
			break;
		}
		if ( fwrite(block,1,amt,gi_fh) != amt ){
			free(file_path_test);
			fclose(gi_fh);
			gzclose(gz_fh);
			return GEOIP_GZIP_READ_ERR;
		}
	}
	gzclose(gz_fh);
	unlink(file_path_gz);
	free(file_path_gz);
	fclose(gi_fh);

	if (verbose == 1)
		GeoIP_printf(f,"Done\n");

	if (verbose == 1) {
    GeoIP_printf(f, WritingFile, GeoIPDBFileName[GEOIP_COUNTRY_EDITION]);
	}

	/* sanity check */
	gi = GeoIP_open(file_path_test, GEOIP_STANDARD);

	if (verbose == 1)
		GeoIP_printf(f,"Performing sanity checks ... ");

	if (gi == NULL) {
		GeoIP_printf(f,"Error opening sanity check database\n");
		return GEOIP_SANITY_OPEN_ERR;
	}

	/* this checks to make sure the files is complete, since info is at the end */
	/* dependent on future databases having MaxMind in info */
	if (verbose == 1)
		GeoIP_printf(f,"database_info  ");
	db_info = GeoIP_database_info(gi);
	if (db_info == NULL) {
		GeoIP_delete(gi);
		if (verbose == 1)
			GeoIP_printf(f,"FAIL\n");
		return GEOIP_SANITY_INFO_FAIL;
	}
	if (strstr(db_info, "MaxMind") == NULL) {
		free(db_info);
		GeoIP_delete(gi);
		if (verbose == 1)
			GeoIP_printf(f,"FAIL\n");
		return GEOIP_SANITY_INFO_FAIL;
	}
	free(db_info);
	if (verbose == 1)
		GeoIP_printf(f,"PASS  ");

	/* this performs an IP lookup test of a US IP address */
	if (verbose == 1)
		GeoIP_printf(f,"lookup  ");
	if (strcmp(GeoIP_country_code_by_addr(gi,"24.24.24.24"), "US") != 0) {
		GeoIP_delete(gi);
		if (verbose == 1)
			GeoIP_printf(f,"FAIL\n");
		return GEOIP_SANITY_LOOKUP_FAIL;
	}
	GeoIP_delete(gi);
	if (verbose == 1)
		GeoIP_printf(f,"PASS\n");

	/* install GeoIP.dat.test -> GeoIP.dat */
	err = rename(file_path_test, GeoIPDBFileName[GEOIP_COUNTRY_EDITION]);
	if (err != 0) {
		GeoIP_printf(f,"GeoIP Install error while renaming file\n");
		return GEOIP_RENAME_ERR;
	}

	if (verbose == 1)
		GeoIP_printf(f,"Done\n");

	return 0;
}

short int GeoIP_update_database_general (char * user_id,char * license_key,char *data_base_type, int verbose,char ** client_ipaddr, void (*f)( char *)) {
	struct hostent *hostlist;
	int sock;
	char * buf, * tmp;
	struct sockaddr_in sa;
	int offset = 0, err;
	char * request_uri;
	char * compr;
	unsigned long comprLen;
	FILE *comp_fh, *cur_db_fh, *gi_fh;
	gzFile gz_fh;
	char * file_path_gz, * file_path_test;
	MD5_CONTEXT context;
	MD5_CONTEXT context2;
	unsigned char buffer[1024], digest[16] ,digest2[16];
	char hex_digest[33] = "0000000000000000000000000000000\0";
	char hex_digest2[33] = "0000000000000000000000000000000\0";
	unsigned int i;
	char *f_str;
	GeoIP * gi;
	char * db_info;
	char *ipaddress;
	char *geoipfilename;
	char *tmpstr;
	int dbtype;
	int lookupresult = 1;
	char block[BLOCK_SIZE];
	int block_size = BLOCK_SIZE;
	size_t len;
	size_t request_uri_len;
	size_t size;

	hostlist = GeoIP_get_host_or_proxy(f);

	if (hostlist == NULL)
		return GEOIP_DNS_ERR;

	if (hostlist->h_addrtype != AF_INET)
		return GEOIP_NON_IPV4_ERR;
	if((sock = socket(AF_INET, SOCK_STREAM, 0)) < 0) {
		return GEOIP_SOCKET_OPEN_ERR;
	}

	memset(&sa, 0, sizeof(struct sockaddr_in));
	sa.sin_port = htons(GeoIPHTTPPort);
	memcpy(&sa.sin_addr, hostlist->h_addr_list[0], hostlist->h_length);
	sa.sin_family = AF_INET;
	
	if (verbose == 1) {
		GeoIP_printf(f,"Connecting to MaxMind GeoIP server\n");
		GeoIP_printf(f, "via Host or Proxy Server: %s:%d\n", hostlist->h_name, GeoIPHTTPPort);
	}
	
	if (connect(sock, (struct sockaddr *)&sa, sizeof(struct sockaddr))< 0)
		return GEOIP_CONNECTION_ERR;
	request_uri = malloc(sizeof(char) * (strlen(GeoIPHTTPRequestFilename) 
                                             + strlen(GeoIPProxyHTTP) + strlen(GeoIPProxiedHost)
                                             + strlen(data_base_type) + strlen(GeoIPUpdateHost) + 1));
	if (request_uri == NULL)
		return GEOIP_OUT_OF_MEMORY_ERR;

	/* get the file name from a web page using the product id */
	sprintf(request_uri,GeoIPHTTPRequestFilename,GeoIPProxyHTTP,GeoIPProxiedHost,data_base_type,GeoIPUpdateHost);
	if (verbose == 1) {
		GeoIP_printf(f, "sending request %s \n",request_uri);
	}

	GeoIP_send_request_uri(sock, request_uri);

	free(request_uri);
	buf = malloc(sizeof(char) * (block_size+4));
	if (buf == NULL)
		return GEOIP_OUT_OF_MEMORY_ERR;
	offset = 0;
	for (;;){
		int amt;
		amt = recv(sock, &buf[offset], block_size,0); 
		if (amt == 0){
			break;
		} else if (amt == -1) {
			free(buf);
			return GEOIP_SOCKET_READ_ERR;
		}
		offset += amt;
		tmp = buf;
		buf = realloc(buf, offset + block_size + 4);
		if ( buf == NULL ){
		    free(tmp);
		    return GEOIP_OUT_OF_MEMORY_ERR;
		}
	}
	buf[offset] = 0;
	offset = 0;
	tmpstr = strstr(buf, "\r\n\r\n");
        if ( tmpstr == NULL ) {
   		free(buf);
		return GEOIP_INVALID_SERVER_RESPONSE;
        }
        /* skip searchstr  "\r\n\r\n" */
        tmpstr += 4;
	if (tmpstr[0] == '.' || strchr(tmpstr, '/') != NULL || strchr(tmpstr, '\\') != NULL) {
		free(buf);
		return GEOIP_INVALID_SERVER_RESPONSE;
	}
	geoipfilename = _GeoIP_full_path_to(tmpstr);
	free(buf);

	/* print the database product id and the database filename */
	if (verbose == 1){
		GeoIP_printf(f, "database product id %s database file name %s \n",data_base_type,geoipfilename);
	}
	_GeoIP_setup_dbfilename();

	/* get MD5 of current GeoIP database file */
	if ((cur_db_fh = fopen (geoipfilename, "rb")) == NULL) {
    GeoIP_printf(f, NoCurrentDB, geoipfilename);
	} else {
		md5_init(&context);
		while ((len = fread (buffer, 1, 1024, cur_db_fh)) > 0)
			md5_write (&context, buffer, len);
		md5_final (&context);
		memcpy(digest,context.buf,16);
		fclose (cur_db_fh);
		for (i = 0; i < 16; i++)
			sprintf (&hex_digest[2*i], "%02x", digest[i]);
    GeoIP_printf(f, MD5Info, hex_digest );
	}
	if (verbose == 1) {
		GeoIP_printf(f,"MD5 sum of database %s is %s \n",geoipfilename,hex_digest);
	}
	if (client_ipaddr[0] == NULL) {
		/* We haven't gotten our IP address yet, so let's request it */
		if ((sock = socket(AF_INET, SOCK_STREAM, 0)) < 0) {
			free(geoipfilename);
			return GEOIP_SOCKET_OPEN_ERR;
		}

		memset(&sa, 0, sizeof(struct sockaddr_in));
		sa.sin_port = htons(GeoIPHTTPPort);
		memcpy(&sa.sin_addr, hostlist->h_addr_list[0], hostlist->h_length);
		sa.sin_family = AF_INET;

		if (verbose == 1)
			GeoIP_printf(f,"Connecting to MaxMind GeoIP Update server\n");

		/* Download gzip file */
		if (connect(sock, (struct sockaddr *)&sa, sizeof(struct sockaddr))< 0) {
			free(geoipfilename);
			return GEOIP_CONNECTION_ERR;
		}
		request_uri = malloc(sizeof(char) * (strlen(GeoIPHTTPRequestClientIP) 
                                                     + strlen(GeoIPProxyHTTP) 
                                                     + strlen(GeoIPProxiedHost)
                                                     + strlen(GeoIPUpdateHost) + 1 ));
		if (request_uri == NULL) {
			free(geoipfilename);
			return GEOIP_OUT_OF_MEMORY_ERR;
		}

		/* get client ip address from MaxMind web page */
		sprintf(request_uri,GeoIPHTTPRequestClientIP,GeoIPProxyHTTP,GeoIPProxiedHost,GeoIPUpdateHost);
		GeoIP_send_request_uri(sock, request_uri);
		if (verbose == 1) {
			GeoIP_printf(f, "sending request %s", request_uri);
		}
		free(request_uri);
		buf = malloc(sizeof(char) * (block_size+1));
		if (buf == NULL) {
			free(geoipfilename);
			return GEOIP_OUT_OF_MEMORY_ERR;
		}
		offset = 0;

		for (;;){
			int amt;
			amt = recv(sock, &buf[offset], block_size,0); 
			if (amt == 0) {
				break;
			} else if (amt == -1) {
				free(buf);
				return GEOIP_SOCKET_READ_ERR;
			}
			offset += amt;
			tmp = buf;
			buf = realloc(buf, offset+block_size+1);
			if ( buf == NULL){
			    free(tmp);
			    return GEOIP_OUT_OF_MEMORY_ERR;
			}
		}

		buf[offset] = 0;
		offset = 0;
		ipaddress = strstr(buf, "\r\n\r\n") + 4; /* get the ip address */
		ipaddress = malloc(strlen(strstr(buf, "\r\n\r\n") + 4)+5);
		strcpy(ipaddress,strstr(buf, "\r\n\r\n") + 4);
		client_ipaddr[0] = ipaddress;
		if (verbose == 1) {
			GeoIP_printf(f, "client ip address: %s\n",ipaddress);
		}
		free(buf);
		close(sock);
	}

	ipaddress = client_ipaddr[0];

	/* make a md5 sum of ip address and license_key and store it in hex_digest2 */
	md5_init(&context2);
	md5_write (&context2, (byte *)license_key, 12);//add license key to the md5 sum
	md5_write (&context2, (byte *)ipaddress, strlen(ipaddress));//add ip address to the md5 sum
	md5_final (&context2);
	memcpy(digest2,context2.buf,16);
	for (i = 0; i < 16; i++)
		snprintf (&hex_digest2[2*i], 3, "%02x", digest2[i]);// change the digest to a hex digest
	if (verbose == 1) {
		GeoIP_printf(f, "md5sum of ip address and license key is %s \n",hex_digest2);
	}

	/* send the request using the user id,product id, 
	 * md5 sum of the prev database and 
	 * the md5 sum of the license_key and ip address */
	if((sock = socket(AF_INET, SOCK_STREAM, 0)) < 0) {
		return GEOIP_SOCKET_OPEN_ERR;
	}
	memset(&sa, 0, sizeof(struct sockaddr_in));
	sa.sin_port = htons(GeoIPHTTPPort);
	memcpy(&sa.sin_addr, hostlist->h_addr_list[0], hostlist->h_length);
	sa.sin_family = AF_INET;
	if (connect(sock, (struct sockaddr *)&sa, sizeof(struct sockaddr))< 0)
		return GEOIP_CONNECTION_ERR;
	request_uri_len = sizeof(char) * 2036;
	request_uri = malloc(request_uri_len);
	if (request_uri == NULL)
	        return GEOIP_OUT_OF_MEMORY_ERR;
	snprintf(request_uri, request_uri_len, GeoIPHTTPRequestMD5,GeoIPProxyHTTP,GeoIPProxiedHost,hex_digest,hex_digest2,user_id,data_base_type);
	GeoIP_send_request_uri(sock, request_uri);
	if (verbose == 1) {
		GeoIP_printf(f, "sending request %s\n",request_uri);
	}

	free(request_uri);

	offset = 0;
	buf = malloc(sizeof(char) * block_size);
	if (buf == NULL)
		return GEOIP_OUT_OF_MEMORY_ERR;

	if (verbose == 1)
		GeoIP_printf(f,"Downloading gzipped GeoIP Database...\n");

	for (;;) {
		int amt;
		amt = recv(sock, &buf[offset], block_size,0);

		if (amt == 0) {
			break;
		} else if (amt == -1) {
			free(buf);
			return GEOIP_SOCKET_READ_ERR;
		}
		offset += amt;
		tmp = buf;
		buf = realloc(buf, offset+block_size);
		if (buf == NULL){
		        free(tmp);
			return GEOIP_OUT_OF_MEMORY_ERR;
		}
	}

	compr = strstr(buf, "\r\n\r\n") + 4;
	comprLen = offset + buf - compr;

	if (strstr(compr, "License Key Invalid") != NULL) {
		if (verbose == 1)
			GeoIP_printf(f,"Failed\n");
		free(buf);
		return GEOIP_LICENSE_KEY_INVALID_ERR;
	} else if (strstr(compr, "No new updates available") != NULL) {
		free(buf);
		GeoIP_printf(f, "%s is up to date, no updates required\n", geoipfilename);
		return GEOIP_NO_NEW_UPDATES;
	} else if (strstr(compr, "Invalid UserId") != NULL){
		free(buf);
		return GEOIP_USER_ID_INVALID_ERR;
	} else if (strstr(compr, "Invalid product ID or subscription expired") != NULL){
		free(buf);
		return GEOIP_PRODUCT_ID_INVALID_ERR;
	}

	if (verbose == 1)
		GeoIP_printf(f,"Done\n");

	GeoIP_printf(f, "Updating %s\n", geoipfilename);

	/* save gzip file */
	file_path_gz = malloc(sizeof(char) * (strlen(geoipfilename) + 4));

	if (file_path_gz == NULL)
		return GEOIP_OUT_OF_MEMORY_ERR;
	strcpy(file_path_gz,geoipfilename);
	strcat(file_path_gz,".gz");
	if (verbose == 1) {
    GeoIP_printf(f, SavingGzip, file_path_gz );
	}
	comp_fh = fopen(file_path_gz, "wb");

	if(comp_fh == NULL) {
		free(file_path_gz);
		free(buf);
		return GEOIP_GZIP_IO_ERR;
	}

	size = fwrite(compr, 1, comprLen, comp_fh);
	fclose(comp_fh);
	free(buf);
        if ( size != comprLen ) {
		return GEOIP_GZIP_IO_ERR;
	}

	if (verbose == 1) {
		GeoIP_printf(f, "download data to a gz file named %s \n",file_path_gz);
		GeoIP_printf(f,"Done\n");
		GeoIP_printf(f,"Uncompressing gzip file ... ");
	}

	file_path_test = malloc(sizeof(char) * (strlen(GeoIPDBFileName[GEOIP_COUNTRY_EDITION]) + 6));
	if (file_path_test == NULL) {
		free(file_path_gz);
		return GEOIP_OUT_OF_MEMORY_ERR;
	}
	strcpy(file_path_test,GeoIPDBFileName[GEOIP_COUNTRY_EDITION]);
	strcat(file_path_test,".test");
	gi_fh = fopen(file_path_test, "wb");
	if(gi_fh == NULL) {
		free(file_path_test);
		free(file_path_gz);
		return GEOIP_TEST_IO_ERR;
	}
	/* uncompress gzip file */
	offset = 0;
	gz_fh = gzopen(file_path_gz, "rb");
	for (;;) {
		int amt;
		amt = gzread(gz_fh, block, block_size);
		if (amt == -1) {
			free(file_path_gz);
			free(file_path_test);
			gzclose(gz_fh);
			fclose(gi_fh);
			return GEOIP_GZIP_READ_ERR;
		}
		if (amt == 0) {
			break;
		}
		if ( amt != fwrite(block,1,amt,gi_fh) ){
			return GEOIP_GZIP_IO_ERR;
		}
	}
	gzclose(gz_fh);
	unlink(file_path_gz);
	free(file_path_gz);
	fclose(gi_fh);

	if (verbose == 1)
		GeoIP_printf(f,"Done\n");

	if (verbose == 1) {
		len = strlen(WritingFile) + strlen(geoipfilename) - 1;
		f_str = malloc(len);
		snprintf(f_str,len,WritingFile,geoipfilename);
		free(f_str);
	}

	/* sanity check */
	gi = GeoIP_open(file_path_test, GEOIP_STANDARD);

	if (verbose == 1)
		GeoIP_printf(f,"Performing sanity checks ... ");

	if (gi == NULL) {
		GeoIP_printf(f,"Error opening sanity check database\n");
		return GEOIP_SANITY_OPEN_ERR;
	}


	/* get the database type */
	dbtype = GeoIP_database_edition(gi);
	if (verbose == 1) {
		GeoIP_printf(f, "Database type is %d\n",dbtype);
	}

	/* this checks to make sure the files is complete, since info is at the end
		 dependent on future databases having MaxMind in info (ISP and Organization databases currently don't have info string */

	if ((dbtype != GEOIP_ISP_EDITION)&&
			(dbtype != GEOIP_ORG_EDITION)) {
		if (verbose == 1)
			GeoIP_printf(f,"database_info  ");
		db_info = GeoIP_database_info(gi);
		if (db_info == NULL) {
			GeoIP_delete(gi);
			if (verbose == 1)
				GeoIP_printf(f,"FAIL null\n");
			return GEOIP_SANITY_INFO_FAIL;
		}
		if (strstr(db_info, "MaxMind") == NULL) {
			free(db_info);
			GeoIP_delete(gi);
			if (verbose == 1)
				GeoIP_printf(f,"FAIL maxmind\n");
			return GEOIP_SANITY_INFO_FAIL;
		}
		free(db_info);
		if (verbose == 1)
			GeoIP_printf(f,"PASS  ");
	}

	/* this performs an IP lookup test of a US IP address */
	if (verbose == 1)
		GeoIP_printf(f,"lookup  ");
	if (dbtype == GEOIP_NETSPEED_EDITION) {
		int netspeed = GeoIP_id_by_name(gi,"24.24.24.24");
		lookupresult = 0;
		if (netspeed == GEOIP_CABLEDSL_SPEED){
			lookupresult = 1;
		}
	}
	if (dbtype == GEOIP_COUNTRY_EDITION) {
		/* if data base type is country then call the function
		 * named GeoIP_country_code_by_addr */
		lookupresult = 1;
		if (strcmp(GeoIP_country_code_by_addr(gi,"24.24.24.24"), "US") != 0) {
			lookupresult = 0;
		}
		if (verbose == 1) {
			GeoIP_printf(f,"testing GEOIP_COUNTRY_EDITION\n");
		}
	}
	if (dbtype == GEOIP_REGION_EDITION_REV1) {
		/* if data base type is region then call the function
		 * named GeoIP_region_by_addr */
		GeoIPRegion *r = GeoIP_region_by_addr(gi,"24.24.24.24");
		lookupresult = 0;
		if (r != NULL) {
			lookupresult = 1;
			free(r);
		}
		if (verbose == 1) {
			GeoIP_printf(f,"testing GEOIP_REGION_EDITION\n");
		}
	}
	if (dbtype == GEOIP_CITY_EDITION_REV1) {
		/* if data base type is city then call the function
		 * named GeoIP_record_by_addr */
		GeoIPRecord *r = GeoIP_record_by_addr(gi,"24.24.24.24");
		lookupresult = 0;
		if (r != NULL) {
			lookupresult = 1;
			free(r);
		}
		if (verbose == 1) {
			GeoIP_printf(f,"testing GEOIP_CITY_EDITION\n");
		}
	}
	if ((dbtype == GEOIP_ISP_EDITION)||
			(dbtype == GEOIP_ORG_EDITION)) {
		/* if data base type is isp or org then call the function
		 * named GeoIP_org_by_addr */
		GeoIPRecord *r = (GeoIPRecord*)GeoIP_org_by_addr(gi,"24.24.24.24");
		lookupresult = 0;
		if (r != NULL) {
			lookupresult = 1;
			free(r);
		}
		if (verbose == 1) {
			if (dbtype == GEOIP_ISP_EDITION) {
				GeoIP_printf(f,"testing GEOIP_ISP_EDITION\n");
			}
			if (dbtype == GEOIP_ORG_EDITION) {
				GeoIP_printf(f,"testing GEOIP_ORG_EDITION\n");
			}
		}
	}
	if (lookupresult == 0) {
		GeoIP_delete(gi);
		if (verbose == 1)
			GeoIP_printf(f,"FAIL\n");
		return GEOIP_SANITY_LOOKUP_FAIL;
	}
	GeoIP_delete(gi);
	if (verbose == 1)
		GeoIP_printf(f,"PASS\n");

	/* install GeoIP.dat.test -> GeoIP.dat */
	err = rename(file_path_test, geoipfilename);
	if (err != 0) {
		GeoIP_printf(f,"GeoIP Install error while renaming file\n");
		return GEOIP_RENAME_ERR;
	}

	if (verbose == 1)
		GeoIP_printf(f,"Done\n");
	free(geoipfilename);
	return 0;
}
