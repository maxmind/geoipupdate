
#include "geoipupdate.h"
#include "md5.h"

#include <ctype.h>
#include <errno.h>
#include <fcntl.h>
#include <getopt.h>
#include <stdarg.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/stat.h>
#include <unistd.h>
#include <zlib.h>

#define ZERO_MD5 ("00000000000000000000000000000000")

static const int ERROR = 1;
static const int OK = 0;

typedef struct {
    char *ptr;
    size_t size;
} in_mem_s;

static int parse_license_file(geoipupdate_s * up);
static int update_country_database(geoipupdate_s * gu);
static void download_to_file(geoipupdate_s * gu, const char *url,
                             const char *fname,
                             char *expected_file_md5);
static int update_database_general_all(geoipupdate_s * gu);
static int update_database_general(geoipupdate_s * gu, const char *product_id);
static in_mem_s *get(geoipupdate_s * gu, const char *url);
static int gunzip_and_replace(geoipupdate_s * gu, const char *gzipfile,
                              const char *geoip_filename,
                              const char *expected_file_md5);

void exit_unless(int expr, const char *fmt, ...)
{
    va_list ap;
    if (expr) {
        return;
    }
    va_start(ap, fmt);
    vfprintf(stderr, fmt, ap);
    va_end(ap);
    exit(1);
}

#define exit_if(expr, ...) exit_unless(!(expr), ## __VA_ARGS__)

void xasprintf(char **ptr, const char *fmt, ...)
{
    va_list ap;
    va_start(ap, fmt);
    int rc = vasprintf(ptr, fmt, ap);
    va_end(ap);
    exit_if(rc == -1, "Out of memory\n");
}

void say_if(int expr, const char *fmt, ...)
{
    va_list ap;
    if (!expr) {
        return;
    }
    va_start(ap, fmt);
    vfprintf(stdout, fmt, ap);
    va_end(ap);
}

#define say(fmt, ...) say_if(1, fmt, ## __VA_ARGS__)

void *xcalloc(size_t nmemb, size_t size)
{
    void *ptr = calloc(nmemb, size);
    exit_if(!ptr, "Out of memory\n");
    return ptr;
}

void *xmalloc(size_t size)
{
    void *ptr = malloc(size);
    exit_if(!ptr, "Out of memory\n");
    return ptr;
}

void *xrealloc(void *ptr, size_t size)
{
    void *mem = realloc(ptr, size);
    exit_if(mem == NULL, "Out of memory\n");
    return mem;
}

static void usage(void)
{
    fprintf(
        stderr,
        "Usage: geoipupdate [-Vhv] [-f license_file] [-d custom directory]\n\n"
        "  -d DIR   store downloaded files in DIR\n"
        "  -f FILE  use configuration found in FILE (see GeoIP.conf(5) man page)\n"
        "  -h       display this help text\n"
        "  -v       use verbose output\n"
        "  -V       display the version and exit\n"
        );
}

int parse_opts(geoipupdate_s * gu, int argc, char *const argv[])
{
    int c;

    opterr = 0;

    while ((c = getopt(argc, argv, "Vvhf:d:")) != -1) {
        switch (c) {
        case 'V':
            puts(PACKAGE_STRING);
            exit(0);
        case 'v':
            gu->verbose = 1;
            break;
        case 'd':
            free(gu->database_dir);
            gu->database_dir = strdup(optarg);
            // The database directory in the config file is ignored if we use -d
            gu->do_not_overwrite_database_directory = 1;
            break;
        case 'f':
            free(gu->license_file);
            gu->license_file = strdup(optarg);
            break;
        case 'h':
        default:
            usage();
            exit(1);
        case '?':
            if (optopt == 'f' || optopt == 'd') {
                fprintf(stderr, "Option -%c requires an argument.\n", optopt);
            } else if (isprint(optopt)) {
                fprintf(stderr, "Unknown option `-%c'.\n", optopt);
            } else{
                fprintf(stderr, "Unknown option character `\\x%x'.\n", optopt);
            }
            exit(1);
        }
    }
    return 0;
}

int main(int argc, char *const argv[])
{
    struct stat st;
    int err = ERROR;
    curl_global_init(CURL_GLOBAL_DEFAULT);
    geoipupdate_s *gu = geoipupdate_s_new();
    if (gu) {
        parse_opts(gu, argc, argv);
        if (parse_license_file(gu)) {
            exit_unless(stat(gu->database_dir, &st) == 0,
                        "%s does not exist\n", gu->database_dir);
            exit_unless(S_ISDIR(st.st_mode), "%s is not a directory\n",
                        gu->database_dir);
            // Note: access(2) checks only the real UID/GID. This is probably
            // okay, but we could perform more complex checks using the stat
            // struct. Alternatively, simply report more thoroughly when we
            // open the file, and avoid potential race issues where permissions
            // change between now and then.
            exit_unless(access(gu->database_dir, W_OK) == 0,
                        "%s is not writable\n", gu->database_dir);
            err = (gu->license.user_id == NO_USER_ID)
                  ? update_country_database(gu)
                  : update_database_general_all(gu);
        }
        geoipupdate_s_delete(gu);
    }
    curl_global_cleanup();
    return err ? ERROR : OK;
}

static ssize_t my_getline(char ** linep, size_t * linecapp, FILE * stream)
{
#if defined HAVE_GETLINE
    return getline(linep, linecapp, stream);
#elif defined HAVE_FGETS
    // Unbelivable, but OS X 10.6 Snow Leopard did not
    // provide getline
    char * p = fgets(*linep, *linecapp, stream);
    return p == NULL ? -1 : strlen(p);
#else
#error Your OS is not supported
#endif
}

static int parse_license_file(geoipupdate_s * up)
{
    say_if(up->verbose, "%s\n", PACKAGE_STRING);
    FILE *fh = fopen(up->license_file, "rb");
    exit_unless(!!fh, "Can't open license file %s\n", up->license_file);
    say_if(up->verbose, "Opened License file %s\n", up->license_file);

    const char *sep = " \t\r\n";
    size_t bsize = 1024;
    char *buffer = (char *)xmalloc(bsize);
    ssize_t read_bytes;
    while ((read_bytes = my_getline(&buffer, &bsize, fh)) != -1) {
        size_t idx = strspn(buffer, sep);
        char *strt = &buffer[idx];
        if (*strt == '#') {
            continue;
        }
        if (sscanf(strt, "UserId %d", &up->license.user_id) == 1) {
            say_if(up->verbose, "UserId %d\n", up->license.user_id);
            continue;
        }
        if (sscanf(strt, "LicenseKey %12s",
                   &up->license.license_key[0]) == 1) {
            say_if(up->verbose, "LicenseKey %s\n", up->license.license_key);
            continue;
        }

        char *p, *last;
        if ((p = strtok_r(strt, sep, &last))) {
            if (!strcmp(p, "ProductIds")) {
                while ((p = strtok_r(NULL, sep, &last))) {
                    product_insert_once(up, p);
                }
            } else if (!strcmp(p, "SkipPeerVerification")) {
                p = strtok_r(NULL, sep, &last);
                exit_if(NULL == p
                        || (0 != strcmp(p, "0") && 0 != strcmp(p, "1")),
                        "SkipPeerVerification must be 0 or 1\n");
                up->skip_peer_verification = atoi(p);
            } else if (!strcmp(p, "Protocol")) {
                p = strtok_r(NULL, sep, &last);
                exit_if(NULL == p || (0 != strcmp(p, "http")
                                      && 0 != strcmp(p, "https")),
                        "Protocol must be http or https\n");
                free(up->proto);
                up->proto = strdup(p);
            } else if (!strcmp(p, "SkipHostnameVerification")) {
                p = strtok_r(NULL, sep, &last);
                exit_if(NULL == p ||
                        (0 != strcmp(p, "0") && 0 != strcmp(p, "1")),
                        "SkipHostnameVerification must be 0 or 1\n");
                up->skip_hostname_verification = atoi(p);
            } else if (!strcmp(p, "Host")) {
                p = strtok_r(NULL, sep, &last);
                exit_if(NULL == p, "Host must be defined\n");
                free(up->host);
                up->host = strdup(p);
            } else if (!strcmp(p, "DatabaseDirectory")) {
                if (!up->do_not_overwrite_database_directory) {
                    p = strtok_r(NULL, sep, &last);
                    exit_if(NULL == p,
                            "DatabaseDirectory must be defined\n");
                    free(up->database_dir);
                    up->database_dir = strdup(p);
                }
            } else if (!strcmp(p, "Proxy")) {
                p = strtok_r(NULL, sep, &last);
                exit_if(NULL == p,
                        "Proxy must be defined 1.2.3.4:12345\n");
                free(up->proxy);
                up->proxy = strdup(p);
            } else if (!strcmp(p, "ProxyUserPassword")) {
                p = strtok_r(NULL, sep, &last);
                exit_if(NULL == p,
                        "ProxyUserPassword must be defined xyz:abc\n");
                free(up->proxy_user_password);
                up->proxy_user_password = strdup(p);
            } else {
                say_if(up->verbose, "Skip unknown directive: %s\n", p);
            }
        }
    }

    free(buffer);
    exit_if(-1 == fclose(fh), "Error closing stream: %s", strerror(errno));
    say_if(up->verbose,
           "Read in license key %s\nNumber of product ids %d\n",
           up->license_file, product_count(up));
    return 1;
}

int md5hex(const char *fname, char *hex_digest)
{
    int bsize = 1024;
    unsigned char buffer[bsize], digest[16];

    size_t len;
    MD5_CONTEXT context;

    FILE *fh = fopen(fname, "rb");
    if (fh == NULL) {
        strcpy(hex_digest, ZERO_MD5);
        return 0;
    }

    struct stat st;
    exit_unless(stat(fname, &st) == 0
                && S_ISREG(st.st_mode), "%s is not a file\n", fname);

    md5_init(&context);
    while ((len = fread(buffer, 1, bsize, fh)) > 0) {
        md5_write(&context, buffer, len);
    }
    md5_final(&context);
    memcpy(digest, context.buf, 16);
    exit_if(-1 == fclose(fh), "Error closing stream: %s", strerror(errno));
    for (int i = 0; i < 16; i++) {
        snprintf(&hex_digest[2 * i], 3, "%02x", digest[i]);
    }
    return 1;
}

static void common_req(CURL * curl, geoipupdate_s * gu)
{
    curl_easy_setopt(curl, CURLOPT_USERAGENT, GEOIP_USERAGENT);
    curl_easy_setopt(curl, CURLOPT_FOLLOWLOCATION, 1);

    // CURLOPT_TCP_KEEPALIVE appeared in 7.25. It is a typedef enum, not a
    // macro so we resort to version detection.
#if LIBCURL_VERSION_NUM >= 0x071900
    curl_easy_setopt(curl, CURLOPT_TCP_KEEPALIVE, 1);
#endif

    if (!strcasecmp(gu->proto, "https")) {
        curl_easy_setopt(curl, CURLOPT_SSL_VERIFYPEER,
                         gu->skip_peer_verification != 0);
        curl_easy_setopt(curl, CURLOPT_SSL_VERIFYHOST,
                         gu->skip_hostname_verification != 0);
    }

    if (gu->proxy_user_password && strlen(gu->proxy_user_password)) {
        say_if(gu->verbose, "Use proxy_user_password: %s\n",
               gu->proxy_user_password);
        curl_easy_setopt(curl, CURLOPT_PROXYUSERPWD, gu->proxy_user_password);
    }
    if (gu->proxy && strlen(gu->proxy)) {
        say_if(gu->verbose, "Use proxy: %s\n", gu->proxy);
        curl_easy_setopt(curl, CURLOPT_PROXY, gu->proxy);
    }
}


size_t get_expected_file_md5(char *buffer, size_t size, size_t nitems,
                             char *md5)
{
    size_t total_size = size * nitems;
    if (strncasecmp(buffer, "X-Database-MD5:", 15) == 0 && total_size > 48) {
        char *start = buffer + 16;
        char *value = start + strspn(start, " \t\r\n");
        strncpy(md5, value, 32);
        md5[32] = '\0';
    }

    return size * nitems;
}

void download_to_file(geoipupdate_s * gu, const char *url, const char *fname,
                      char *expected_file_md5)
{
    FILE *f = fopen(fname, "wb");
    if (NULL == f) {
        fprintf(stderr, "Can't open %s: %s\n", fname, strerror(errno));
        exit(1);
    }

    say_if(gu->verbose, "url: %s\n", url);
    CURL *curl = gu->curl;

    expected_file_md5[0] = '\0';
    curl_easy_setopt(curl, CURLOPT_HEADERFUNCTION, get_expected_file_md5);
    curl_easy_setopt(curl, CURLOPT_HEADERDATA, expected_file_md5);

    curl_easy_setopt(curl, CURLOPT_WRITEFUNCTION, NULL);
    curl_easy_setopt(curl, CURLOPT_WRITEDATA, (void *)f);

    curl_easy_setopt(curl, CURLOPT_URL, url);
    common_req(curl, gu);
    CURLcode res = curl_easy_perform(curl);

    exit_unless(res == CURLE_OK,
                "curl_easy_perform() failed: %s\nConnect to %s\n",
                curl_easy_strerror(res), url);

    long status = 0;
    curl_easy_getinfo(curl, CURLINFO_RESPONSE_CODE, &status);

    exit_if( status < 200 || status >= 300,
             "Received an unexpected HTTP status code of %ld from %s\n",
             status, url);

    exit_if(-1 == fclose(f), "Error closing stream: %s", strerror(errno));
}

static size_t mem_cb(void *contents, size_t size, size_t nmemb, void *userp)
{
    size_t realsize = size * nmemb;

    if (realsize == 0) {
        return realsize;
    }

    in_mem_s *mem = (in_mem_s *)userp;

    mem->ptr = (char *)xrealloc(mem->ptr, mem->size + realsize + 1);
    memcpy(&(mem->ptr[mem->size]), contents, realsize);
    mem->size += realsize;
    mem->ptr[mem->size] = 0;

    return realsize;
}

static in_mem_s *in_mem_s_new(void)
{
    in_mem_s *mem = (in_mem_s *)xmalloc(sizeof(in_mem_s));
    mem->ptr = (char *)xcalloc(1, 1);
    mem->size = 0;
    return mem;
}

static void in_mem_s_delete(in_mem_s * mem)
{
    if (mem) {
        free(mem->ptr);
        free(mem);
    }
}

static in_mem_s *get(geoipupdate_s * gu, const char *url)
{
    in_mem_s *mem = in_mem_s_new();

    say_if(gu->verbose, "url: %s\n", url);
    CURL *curl = gu->curl;
    curl_easy_setopt(curl, CURLOPT_URL, url);
    curl_easy_setopt(curl, CURLOPT_WRITEFUNCTION, mem_cb);
    curl_easy_setopt(curl, CURLOPT_WRITEDATA, (void *)mem);
    common_req(curl, gu);
    CURLcode res = curl_easy_perform(curl);
    exit_unless(res == CURLE_OK,
                "curl_easy_perform() failed: %s\nConnect to %s\n",
                curl_easy_strerror(res), url);

    long status = 0;
    curl_easy_getinfo(curl, CURLINFO_RESPONSE_CODE, &status);

    exit_if( status < 200 || status >= 300,
             "Received an unexpected HTTP status code of %ld from %s",
             status, url);

    return mem;
}

void md5hex_license_ipaddr(geoipupdate_s * gu, const char *client_ipaddr,
                           char *new_digest_str)
{
    unsigned char digest[16];
    MD5_CONTEXT context;
    md5_init(&context);
    md5_write(&context, (unsigned char *)gu->license.license_key,
              strlen(gu->license.license_key));
    md5_write(&context, (unsigned char *)client_ipaddr, strlen(client_ipaddr));
    md5_final(&context);
    memcpy(digest, context.buf, 16);
    for (int i = 0; i < 16; i++) {
        snprintf(&new_digest_str[2 * i], 3, "%02x", digest[i]);
    }
}

static int update_database_general(geoipupdate_s * gu, const char *product_id)
{
    char *url, *geoip_filename, *geoip_gz_filename, *client_ipaddr;
    char hex_digest[33], hex_digest2[33];

    xasprintf(&url, "%s://%s/app/update_getfilename?product_id=%s",
              gu->proto, gu->host, product_id);
    in_mem_s *mem = get(gu, url);
    free(url);
    if (mem->size == 0) {
        fprintf(stderr, "product_id %s not found\n", product_id);
        in_mem_s_delete(mem);
        return ERROR;
    }
    xasprintf(&geoip_filename, "%s/%s", gu->database_dir, mem->ptr);
    in_mem_s_delete(mem);
    md5hex(geoip_filename, hex_digest);
    say_if(gu->verbose, "md5hex_digest: %s\n", hex_digest);
    xasprintf(&url, "%s://%s/app/update_getipaddr", gu->proto, gu->host);
    mem = get(gu, url);
    free(url);
    client_ipaddr = strdup(mem->ptr);
    in_mem_s_delete(mem);

    say_if(gu->verbose, "Client IP address: %s\n", client_ipaddr);
    md5hex_license_ipaddr(gu, client_ipaddr, hex_digest2);
    free(client_ipaddr);
    say_if(gu->verbose, "md5hex_digest2: %s\n", hex_digest2);

    xasprintf(
        &url,
        "%s://%s/app/update_secure?db_md5=%s&challenge_md5=%s&user_id=%d&edition_id=%s",
        gu->proto, gu->host, hex_digest, hex_digest2,
        gu->license.user_id, product_id);
    xasprintf(&geoip_gz_filename, "%s.gz", geoip_filename);

    char expected_file_md5[33];
    download_to_file(gu, url, geoip_gz_filename, expected_file_md5);
    free(url);
    int rc = gunzip_and_replace(gu, geoip_gz_filename, geoip_filename,
                                expected_file_md5);
    free(geoip_gz_filename);
    free(geoip_filename);
    return rc;
}

static int update_database_general_all(geoipupdate_s * gu)
{
    int err = 0;
    for (product_s ** next = &gu->license.first; *next; next =
             &(*next)->next) {
        err |= update_database_general(gu, (*next)->product_id);
    }
    return err;
}

static int update_country_database(geoipupdate_s * gu)
{
    char *geoip_filename, *geoip_gz_filename, *url;
    char hex_digest[33];
    xasprintf(&geoip_filename, "%s/GeoIP.dat", gu->database_dir);
    xasprintf(&geoip_gz_filename, "%s/GeoIP.dat.gz", gu->database_dir);

    md5hex(geoip_filename, hex_digest);
    say_if(gu->verbose, "md5hex_digest: %s\n", hex_digest);
    xasprintf(&url,
              "%s://%s/app/update?license_key=%s&md5=%s",
              gu->proto, gu->host, &gu->license.license_key[0], hex_digest);

    char expected_file_md5[33];
    download_to_file(gu, url, geoip_gz_filename, expected_file_md5);
    free(url);

    int rc = gunzip_and_replace(gu, geoip_gz_filename, geoip_filename,
                                expected_file_md5);

    free(geoip_gz_filename);
    free(geoip_filename);
    return rc ? ERROR : OK;
}

static int gunzip_and_replace(geoipupdate_s * gu, const char *gzipfile,
                              const char *geoip_filename,
                              const char *expected_file_md5)
{
    gzFile gz_fh;
    FILE *fh = fopen(gzipfile, "rb");
    exit_if(NULL == fh, "Can't open %s\n", gzipfile);
    size_t bsize = 8096;
    char *buffer = (char *)xmalloc(bsize);
    ssize_t read_bytes = my_getline(&buffer, &bsize, fh);
    exit_if(-1 == fclose(fh), "Error closing stream: %s", strerror(errno));
    if (read_bytes < 0) {
        fprintf(stderr, "Read error %s\n", gzipfile);
        unlink(gzipfile);
        free(buffer);
        return ERROR;
    }
    const char *no_new_upd = "No new updates available";
    if (!strncmp(no_new_upd, buffer, strlen(no_new_upd))) {
        say_if(gu->verbose, "%s\n", no_new_upd);
        unlink(gzipfile);
        free(buffer);
        return OK;
    }
    if (strncmp(buffer, "\x1f\x8b", 2)) {
        // error not a zip file
        unlink(gzipfile);
        printf("%s is not a valid gzip file\n", gzipfile);
        return ERROR;
    }

    // We do this here as we have to check that there is an update before
    // we check for the header.
    exit_unless( 32 == strnlen(expected_file_md5, 33),
                 "Did not receive a valid expected database MD5 from server\n");

    char *file_path_test;
    xasprintf(&file_path_test, "%s.test", geoip_filename);
    say_if(gu->verbose, "Uncompress file %s to %s\n", gzipfile, file_path_test);
    gz_fh = gzopen(gzipfile, "rb");
    exit_if(gz_fh == NULL, "Can't open %s\n", gzipfile);
    FILE *fhw = fopen(file_path_test, "wb");
    exit_if(fhw == NULL, "Can't open %s\n", file_path_test);

    for (;; ) {
        int amt = gzread(gz_fh, buffer, bsize);
        if (amt == 0) {
            break;              // EOF
        }
        exit_if(amt == -1, "Gzip read error while reading from %s\n", gzipfile);
        exit_unless(fwrite(buffer, 1, amt, fhw) == (size_t)amt,
                    "Gzip write error\n");
    }
    exit_if(-1 == fclose(fhw), "Error closing stream: %s", strerror(errno));
    exit_if(gzclose(gz_fh) != Z_OK, "Gzip read error while closing from %s\n",
            gzipfile);
    free(buffer);

    char actual_md5[33];
    md5hex(file_path_test, actual_md5);
    exit_if(strncasecmp(actual_md5, expected_file_md5, 32),
            "MD5 of new database (%s) does not match expected MD5 (%s)",
            actual_md5, expected_file_md5);

    say_if(gu->verbose, "Rename %s to %s\n", file_path_test, geoip_filename);
    int err = rename(file_path_test, geoip_filename);
    exit_if(err, "Rename %s to %s failed\n", file_path_test, geoip_filename);

    // fsync directory to ensure the rename is durable
    int dirfd = open(gu->database_dir, O_DIRECTORY);
    exit_if(-1 == dirfd, "Error opening database directory: %s",
            strerror(errno));
    exit_if(-1 == fsync(dirfd), "Error syncing database directory: %s",
            strerror(errno));
    exit_if(-1 == close(dirfd), "Error closing database directory: %s",
            strerror(errno));
    exit_if(-1 == unlink(gzipfile), "Error unlinking %s: %s", gzipfile,
            strerror(errno));

    free(file_path_test);
    return OK;
}
