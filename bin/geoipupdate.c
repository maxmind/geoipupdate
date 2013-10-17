
#include "geoipupdate.h"
#include <curl/curl.h>

#include <unistd.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <getopt.h>
#include <ctype.h>
#include <stdarg.h>
#include "md5.h"
#include <zlib.h>
#include <sys/stat.h>

typedef struct {
    char *ptr;
    size_t size;
} in_mem_s;

int parse_license_file(geoipupdate_s * up);
void update_country_database(geoipupdate_s * gu);
static void get_to_disc(geoipupdate_s * gu, const char *url, const char *fname);
static void update_database_general_all(geoipupdate_s * gu);
static void update_database_general(geoipupdate_s * gu, const char *product_id);
static in_mem_s *get(geoipupdate_s * gu, const char *url);
static void gunzip_and_replace(geoipupdate_s * gu, const char *gzipfile,
                               const char *geoip_filename);

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
    curl_global_init(CURL_GLOBAL_DEFAULT);
    geoipupdate_s *gu = geoipupdate_s_new();
    if (gu) {
        parse_opts(gu, argc, argv);
        if (parse_license_file(gu)) {
            exit_unless(stat(gu->database_dir, &st) == 0,
                        "%s does not exisits\n", gu->database_dir);
            exit_unless(S_ISDIR(st.st_mode), "%s is not a directory\n",
                        gu->database_dir);
            if (gu->license.user_id == NO_USER_ID) {
                update_country_database(gu);
            } else{
                update_database_general_all(gu);
            }
        }
        geoipupdate_s_delete(gu);
    }
    curl_global_cleanup();
    return 0;
}

int parse_license_file(geoipupdate_s * up)
{
    FILE *fh = fopen(up->license_file, "r");
    exit_unless(!!fh, "Can't open license file %s\n", up->license_file);
    say_if(up->verbose, "Opened License file %s\n", up->license_file);

    const char *sep = " \t\r\n";
    size_t bsize = 1024;
    char *buffer = (char *)xmalloc(bsize);
    ssize_t read_bytes;
    while ((read_bytes = getline(&buffer, &bsize, fh)) != -1) {
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
                exit_unless(p != NULL
                            && (!strcmp(p, "0") || !strcmp(p, "1")),
                            "SkipPeerVerification must be 0 or 1\n");
                up->skip_peer_verification = atoi(p);
            } else if (!strcmp(p, "Protocol")) {
                p = strtok_r(NULL, sep, &last);
                exit_unless(p != NULL && (!strcmp(p, "http")
                                          || !strcmp(p, "https")),
                            "Protocol must be http or https\n");
                free(up->proto);
                up->proto = strdup(p);
            } else if (!strcmp(p, "SkipHostnameVerification")) {
                p = strtok_r(NULL, sep, &last);
                exit_unless(p != NULL
                            && (!strcmp(p, "0") || !strcmp(p, "1")),
                            "SkipHostnameVerification must be 0 or 1\n");
                up->skip_hostname_verification = atoi(p);
            } else if (!strcmp(p, "Host")) {
                p = strtok_r(NULL, sep, &last);
                exit_unless(p != NULL, "Host must be defined\n");
                free(up->host);
                up->host = strdup(p);
            } else if (!strcmp(p, "DatabaseDirectory")) {
                if (!up->do_not_overwrite_database_directory) {
                    p = strtok_r(NULL, sep, &last);
                    exit_unless(p != NULL,
                                "DatabaseDirectory must be defined\n");
                    free(up->database_dir);
                    up->database_dir = strdup(p);
                }
            } else if (!strcmp(p, "Proxy")) {
                p = strtok_r(NULL, sep, &last);
                exit_unless(p != NULL,
                            "Proxy must be defined 1.2.3.4:12345\n");
                free(up->proxy);
                up->proxy = strdup(p);
            } else if (!strcmp(p, "ProxyUserPassword")) {
                p = strtok_r(NULL, sep, &last);
                exit_unless(p != NULL,
                            "ProxyUserPassword must be defined xyz:abc\n");
                free(up->proxy_user_password);
                up->proxy_user_password = strdup(p);
            }
        }
    }

    free(buffer);
    fclose(fh);
    say_if(up->verbose,
           "Read in license key %s\nNumber of product ids %d\n",
           up->license_file, product_count(up));
    return 1;
}

int md5hex(const char *fname, char *hex_digest)
{
    int bsize = 1024;
    unsigned char buffer[bsize], digest[16];
    const char zero_hex_digest[34] = "00000000000000000000000000000000\0";
    size_t len;
    MD5_CONTEXT context;

    FILE *fh = fopen(fname, "rb");
    if (fh == NULL) {
        strcpy(hex_digest, zero_hex_digest);
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
    fclose(fh);
    for (int i = 0; i < 16; i++) {
        snprintf(&hex_digest[2 * i], 3, "%02x", digest[i]);
    }
    return 1;
}

static void common_req(CURL * curl, geoipupdate_s * gu)
{
    curl_easy_setopt(curl, CURLOPT_USERAGENT, GEOIP_USERAGENT);
    curl_easy_setopt(curl, CURLOPT_FOLLOWLOCATION, 1);
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

void get_to_disc(geoipupdate_s * gu, const char *url, const char *fname)
{
    FILE *f = fopen(fname, "w");
    exit_unless(f != NULL, "Can't open %s\n", fname);
    say_if(gu->verbose, "url: %s\n", url);
    CURL *curl = curl_easy_init();
    curl_easy_setopt(curl, CURLOPT_WRITEFUNCTION, NULL);
    curl_easy_setopt(curl, CURLOPT_WRITEDATA, (void *)f);
    curl_easy_setopt(curl, CURLOPT_URL, url);
    common_req(curl, gu);
    CURLcode res = curl_easy_perform(curl);

    exit_unless(res == CURLE_OK,
                "curl_easy_perform() failed: %s\nConnect to %s\n",
                curl_easy_strerror(res), url);

    curl_easy_cleanup(curl);
    fclose(f);
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
    CURL *curl = curl_easy_init();
    curl_easy_setopt(curl, CURLOPT_URL, url);
    curl_easy_setopt(curl, CURLOPT_WRITEFUNCTION, mem_cb);
    curl_easy_setopt(curl, CURLOPT_WRITEDATA, (void *)mem);
    common_req(curl, gu);
    CURLcode res = curl_easy_perform(curl);
    exit_unless(res == CURLE_OK,
                "curl_easy_perform() failed: %s\nConnect to %s\n",
                curl_easy_strerror(res), url);
    curl_easy_cleanup(curl);
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

static void update_database_general(geoipupdate_s * gu, const char *product_id)
{
    char *url, *geoip_filename, *geoip_gz_filename, *client_ipaddr;
    char hex_digest[33], hex_digest2[33];

    xasprintf(&url, "%s://%s/app/update_getfilename?product_id=%s",
              gu->proto, gu->host, product_id);
    in_mem_s *mem = get(gu, url);
    free(url);
    exit_if(mem->size == 0, "product_id %s not found\n", product_id);
    xasprintf(&geoip_filename, "%s/%s", gu->database_dir, mem->ptr);
    in_mem_s_delete(mem);
    md5hex(geoip_filename, hex_digest);
    say_if(gu->verbose, "md5hex_digest: %s\n", hex_digest);
    xasprintf(&url, "%s://%s/app/update_getipaddr", gu->proto, gu->host);
    mem = get(gu, url);
    free(url);
    client_ipaddr = strdup(mem->ptr);
    in_mem_s_delete(mem);

    md5hex_license_ipaddr(gu, client_ipaddr, hex_digest2);
    free(client_ipaddr);
    say_if(gu->verbose, "md5hex_digest2: %s\n", hex_digest2);

    xasprintf(
        &url,
        "%s://%s/app/update_secure?db_md5=%s&challenge_md5=%s&user_id=%d&edition_id=%s",
        gu->proto, gu->host, hex_digest, hex_digest2,
        gu->license.user_id, product_id);
    xasprintf(&geoip_gz_filename, "%s.gz", geoip_filename);
    get_to_disc(gu, url, geoip_gz_filename);
    free(url);
    gunzip_and_replace(gu, geoip_gz_filename, geoip_filename);
    free(geoip_gz_filename);
    free(geoip_filename);
}

static void update_database_general_all(geoipupdate_s * gu)
{
    for (product_s ** next = &gu->license.first; *next; next =
             &(*next)->next) {
        update_database_general(gu, (*next)->product_id);
    }
}

void update_country_database(geoipupdate_s * gu)
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
    get_to_disc(gu, url, geoip_gz_filename);
    free(url);

    gunzip_and_replace(gu, geoip_gz_filename, geoip_filename);
    free(geoip_gz_filename);
    free(geoip_filename);
}

static void gunzip_and_replace(geoipupdate_s * gu, const char *gzipfile,
                               const char *geoip_filename)
{
    gzFile gz_fh;
    FILE *fh = fopen(gzipfile, "rb");
    exit_unless(fh != NULL, "Can't open %s\n", gzipfile);
    size_t bsize = 8096;
    char *buffer = (char *)xmalloc(bsize);
    ssize_t read_bytes = getline(&buffer, &bsize, fh);
    fclose(fh);
    exit_unless(read_bytes >= 0, "Read error %s\n", gzipfile);
    const char *no_new_upd = "No new updates available";
    if (!strncmp(no_new_upd, buffer, strlen(no_new_upd))) {
        say_if(gu->verbose, "%s\n", no_new_upd);
        unlink(gzipfile);
        free(buffer);
        return;
    }
    if (strncmp(buffer, "\x1f\x8b", 2)) {
        // error not a zip file
        unlink(gzipfile);
        exit_unless(0, "%s\n", buffer);
    }
    char *file_path_test;
    xasprintf(&file_path_test, "%s.test", geoip_filename);
    say_if(gu->verbose, "Uncompress file %s to %s\n", gzipfile, file_path_test);
    gz_fh = gzopen(gzipfile, "rb");
    exit_unless(gz_fh != NULL, "Can't open %s\n", gzipfile);
    FILE *fhw = fopen(file_path_test, "wb");
    exit_unless(fhw >= 0, "Can't open %s\n", file_path_test);

    for (;; ) {
        int amt = gzread(gz_fh, buffer, bsize);
        if (amt == 0) {
            break;              // EOF
        }
        exit_if(amt == -1, "Gzip read error while reading from %s\n", gzipfile);
        exit_unless(fwrite(buffer, 1, amt, fhw) == (size_t)amt,
                    "Gzip write error\n");
    }
    fclose(fhw);
    gzclose(gz_fh);
    free(buffer);
    say_if(gu->verbose, "Rename %s to %s\n", file_path_test, geoip_filename);
    int err = rename(file_path_test, geoip_filename);
    exit_if(err, "Rename %s to %s failed\n", file_path_test, geoip_filename);
    unlink(gzipfile);
    free(file_path_test);
}
