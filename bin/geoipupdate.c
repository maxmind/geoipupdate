#include "functions.h"
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
#include <sys/types.h>
#include <unistd.h>
#include <zlib.h>

#define ZERO_MD5 ("00000000000000000000000000000000")
#define say(fmt, ...) say_if(1, fmt, ## __VA_ARGS__)

enum gu_status {
    GU_OK = 0,
    GU_ERROR = 1,
};

typedef struct {
    char *ptr;
    size_t size;
} in_mem_s;

static void xasprintf(char **, const char *, ...);
static void *xcalloc(size_t, size_t);
static void *xrealloc(void *, size_t);
static void usage(void);
static int parse_opts(geoipupdate_s *, int, char *const []);
static ssize_t my_getline(char **, size_t *, FILE *);
static int parse_license_file(geoipupdate_s *);
static char * join_path(char const * const, char const * const);
static int acquire_run_lock(geoipupdate_s const * const);
static int md5hex(const char *, char *);
static void common_req(CURL *, geoipupdate_s *);
static size_t get_expected_file_md5(char *, size_t, size_t,
                                    char *);
static void download_to_file(geoipupdate_s *, const char *,
                             const char *, char *);
static size_t mem_cb(void *, size_t, size_t, void *);
static in_mem_s *in_mem_s_new(void);
static void in_mem_s_delete(in_mem_s *);
static in_mem_s *get(geoipupdate_s *, const char *);
static void md5hex_license_ipaddr(geoipupdate_s *, const char *,
                                  char *);
static int update_database_general_all(geoipupdate_s *);
static int update_database_general(geoipupdate_s *, const char *);
static int update_country_database(geoipupdate_s *);
static int gunzip_and_replace(geoipupdate_s const * const,
                              char const * const, char const * const,
                              char const * const);

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

static void xasprintf(char **ptr, const char *fmt, ...)
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

static void *xcalloc(size_t nmemb, size_t size)
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

static void *xrealloc(void *ptr, size_t size)
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

static int parse_opts(geoipupdate_s * gu, int argc, char *const argv[])
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
            exit_if(NULL == gu->database_dir,
                    "Unable to allocate memory for database directory path.");

            // The database directory in the config file is ignored if we use -d
            gu->do_not_overwrite_database_directory = 1;
            break;
        case 'f':
            free(gu->license_file);
            gu->license_file = strdup(optarg);
            exit_if(NULL == gu->license_file,
                    "Unable to allocate memory for license file path.\n");
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
    return GU_OK;
}

int main(int argc, char *const argv[])
{
    struct stat st;
    int err = GU_ERROR;
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

            if (acquire_run_lock(gu) != 0) {
                geoipupdate_s_delete(gu);
                curl_global_cleanup();
                return GU_ERROR;
            }

            err = (gu->license.user_id == NO_USER_ID)
                  ? update_country_database(gu)
                  : update_database_general_all(gu);
        }
        geoipupdate_s_delete(gu);
    }
    curl_global_cleanup();
    return err ? GU_ERROR : GU_OK;
}

static ssize_t my_getline(char ** linep, size_t * linecapp, FILE * stream)
{
#if defined HAVE_GETLINE
    return getline(linep, linecapp, stream);
#elif defined HAVE_FGETS
    // Unbelievable, but OS X 10.6 Snow Leopard did not provide getline
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
                exit_if(NULL == up->proto,
                        "Unable to allocate memory for request protocol.\n");
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
                exit_if(NULL == up->host,
                        "Unable to allocate memory for update host.\n");
            } else if (!strcmp(p, "DatabaseDirectory")) {
                if (!up->do_not_overwrite_database_directory) {
                    p = strtok_r(NULL, sep, &last);
                    exit_if(NULL == p,
                            "DatabaseDirectory must be defined\n");
                    free(up->database_dir);
                    up->database_dir = strdup(p);
                    exit_if(
                        NULL == up->database_dir,
                        "Unable to allocate memory for database directory path.\n");
                }
            } else if (!strcmp(p, "Proxy")) {
                p = strtok_r(NULL, sep, &last);
                exit_if(NULL == p,
                        "Proxy must be defined 1.2.3.4:12345\n");
                free(up->proxy);
                up->proxy = strdup(p);
                exit_if(NULL == up->proxy,
                        "Unable to allocate memory for proxy host.\n");
            } else if (!strcmp(p, "ProxyUserPassword")) {
                p = strtok_r(NULL, sep, &last);
                exit_if(NULL == p,
                        "ProxyUserPassword must be defined xyz:abc\n");
                free(up->proxy_user_password);
                up->proxy_user_password = strdup(p);
                exit_if(NULL == up->proxy_user_password,
                        "Unable to allocate memory for proxy credentials.\n");
            } else if (!strcmp(p, "LockFile")) {
                p = strtok_r(NULL, sep, &last);
                exit_if(NULL == p,
                        "LockFile must be a file path\n");
                // We could check the value looks like a path, but trying to use
                // it will fail if it isn't.
                free(up->lock_file);
                up->lock_file = strdup(p);
                exit_if(NULL == up->lock_file,
                        "Out of memory allocating LockFile string\n");
            } else {
                say_if(up->verbose, "Skip unknown directive: %s\n", p);
            }
        }
    }

    // If we don't have a LockFile specified, then default to .geoipupdate.lock
    // in the database directory. Do this here as the database directory may
    // have been set either on the command line or in the configuration file.
    if (strlen(up->lock_file) == 0) {
        free(up->lock_file);
        up->lock_file = join_path(up->database_dir, ".geoipupdate.lock");
        exit_if(NULL == up->lock_file, "Unable to create path to lock file.");
    }

    free(buffer);
    exit_if(-1 == fclose(fh), "Error closing stream: %s", strerror(errno));
    say_if(up->verbose,
           "Read in license key %s\nNumber of product ids %d\n",
           up->license_file, product_count(up));
    return 1;
}

// Given a directory and a filename in that directory, combine the two to make a
// path to the file.
//
// TODO: This function assumes Unix style paths (/ separator).
//
// TODO: This function performs no validation on the given inputs beyond that
// they are present.
static char * join_path(char const * const dir, char const * const file)
{
    size_t sz = -1;
    char * path = NULL;

    if (dir == NULL || strlen(dir) == 0 || file == NULL || strlen(file) == 0) {
        fprintf(stderr, "join_path: %s\n", strerror(EINVAL));
        return NULL;
    }

    // dir '/' file '\0'
    sz = strlen(dir) + 1 + strlen(file) + 1;

    path = calloc(sz, sizeof(char));
    if (path == NULL) {
        fprintf(stderr, "join_path: %s\n", strerror(errno));
        return NULL;
    }

    strcat(path, dir);
    strcat(path, "/");
    strcat(path, file);

    return path;
}

// Acquire a lock to ensure this is the only running geoipupdate instance. This
// is to avoid race conditions where multiple geoipupdate instances run at
// once, leading to failures.
//
// Wait for a lock. If locking fails, return non-zero. If it succeeds, return
// zero.
//
// Use fcntl(2) to acquire the lock. The primary rationale to use this over
// something like open(2) with O_EXCL is that we don't need to perform clean up
// to release the lock. In particular, if execution ends unexpectedly, such as
// due to a crash, the lock will be automatically released. It also means we
// don't need to worry about lock bookkeeping even in the normal case, since
// the lock gets released automatically at program exit.
//
// This method does have the drawback that removing the lock file is not
// possible due to the potential for race conditions. Consider the case where
// another instance opens the lock file, then we remove the file and close the
// file (releasing our lock), then that other instance acquires a lock. At the
// same time, another instance runs and creates the file anew and also acquires
// a lock.
static int acquire_run_lock(geoipupdate_s const * const gu)
{
    int fd = -1;
    struct flock fl;
    int i = 0;

    memset(&fl, 0, sizeof(struct flock));

    if (gu == NULL || gu->lock_file == NULL || strlen(gu->lock_file) == 0) {
        fprintf(stderr, "maybe_acquire_run_lock: %s\n", strerror(EINVAL));
        return 1;
    }

    fd = open(gu->lock_file, O_WRONLY | O_CREAT, S_IRUSR | S_IWUSR);
    if (fd == -1) {
        fprintf(stderr, "Unable to open lock file %s: %s\n", gu->lock_file,
                strerror(errno));
        return 1;
    }

    fl.l_type = F_WRLCK;

    // Try 3 times to acquire a lock. Arbitrary number.
    for (i = 0; i < 3; i++) {
        if (fcntl(fd, F_SETLKW, &fl) == 0) {
            // Locked.
            return 0;
        }

        // Interrupted? Retry.
        if (errno == EINTR) {
            continue;
        }

        // Something else went wrong. Abort.
        fprintf(stderr, "Unable to acquire lock on %s: %s\n", gu->lock_file,
                strerror(errno));
        close(fd);
        return 1;
    }

    fprintf(stderr, "Unable to acquire lock on %s: Gave up after %d attempts\n",
            gu->lock_file, i);
    close(fd);
    return 1;
}

static int md5hex(const char *fname, char *hex_digest)
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

static size_t get_expected_file_md5(char *buffer, size_t size, size_t nitems,
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

// Make an HTTP request and download the response body to a file.
//
// If the HTTP status is not 2xx, we have a error message in the body rather
// than a file. Write it to stderr and exit.
//
// TODO(wstorey@maxmind.com): Return boolean/int whether we succeeded rather
// than exiting. Beyond being cleaner and easier to test, it will allow us to
// clean up after ourselves better.
static void download_to_file(geoipupdate_s * gu, const char *url,
                             const char *fname, char *expected_file_md5)
{
    FILE *f = fopen(fname, "wb");
    if (f == NULL) {
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

    if (fclose(f) == -1) {
        fprintf(stderr, "Error closing file: %s: %s\n", fname, strerror(errno));
        unlink(fname);
        exit(1);
    }

    if (status < 200 || status >= 300) {
        fprintf(stderr,
                "Received an unexpected HTTP status code of %ld from %s:\n",
                status,
                url);
        // The response should contain a message containing exactly why.
        char * const message = slurp_file(fname);
        if (message) {
            fprintf(stderr, "%s\n", message);
            free(message);
        }
        unlink(fname);
        exit(1);
    }

    // We have HTTP 2xx.

    // In this case, the server must have told us the current MD5 hash of the
    // database we asked for.
    if (strnlen(expected_file_md5, 33) != 32) {
        fprintf(stderr,
                "Did not receive a valid expected database MD5 from server\n");
        unlink(fname);
        exit(1);
    }
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

// Generate an MD5 hash of the concatenation of license key with IP address.
//
// This hash is suitable for the challenge parameter for downloading from the
// /update_secure endpoint.
static void md5hex_license_ipaddr(geoipupdate_s * gu, const char *client_ipaddr,
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
    char *url = NULL, *geoip_filename = NULL, *geoip_gz_filename = NULL,
    *client_ipaddr = NULL;
    char hex_digest[33] = { 0 }, hex_digest2[33] = { 0 };

    // Get the filename.
    xasprintf(&url, "%s://%s/app/update_getfilename?product_id=%s",
              gu->proto, gu->host, product_id);
    in_mem_s *mem = get(gu, url);
    free(url);
    if (mem->size == 0) {
        fprintf(stderr, "product_id %s not found\n", product_id);
        in_mem_s_delete(mem);
        return GU_ERROR;
    }
    xasprintf(&geoip_filename, "%s/%s", gu->database_dir, mem->ptr);
    in_mem_s_delete(mem);

    // Calculate the MD5 hash of the database we currently have, if any. We get
    // back a zero MD5 hash if we don't have it yet.
    md5hex(geoip_filename, hex_digest);
    say_if(gu->verbose, "md5hex_digest: %s\n", hex_digest);

    // Look up our IP.
    xasprintf(&url, "%s://%s/app/update_getipaddr", gu->proto, gu->host);
    mem = get(gu, url);
    free(url);

    client_ipaddr = strdup(mem->ptr);
    if (NULL == client_ipaddr) {
        fprintf(stderr, "Unable to allocate memory for client IP address.\n");
        free(geoip_filename);
        in_mem_s_delete(mem);
        return GU_ERROR;
    }

    in_mem_s_delete(mem);

    say_if(gu->verbose, "Client IP address: %s\n", client_ipaddr);

    // Make the challenge MD5 hash.
    md5hex_license_ipaddr(gu, client_ipaddr, hex_digest2);

    free(client_ipaddr);
    say_if(gu->verbose, "md5hex_digest2 (challenge): %s\n", hex_digest2);

    // Download.
    xasprintf(
        &url,
        "%s://%s/app/update_secure?db_md5=%s&challenge_md5=%s&user_id=%d&edition_id=%s",
        gu->proto, gu->host, hex_digest, hex_digest2,
        gu->license.user_id, product_id);
    xasprintf(&geoip_gz_filename, "%s.gz", geoip_filename);

    char expected_file_md5[33] = { 0 };
    download_to_file(gu, url, geoip_gz_filename, expected_file_md5);
    free(url);

    // Was there actually an update? We can tell because if not we will have
    // the same MD5 reported back. Note in the past we would check the response
    // body which does still say whether we have an update.
    if (strcmp(hex_digest, expected_file_md5) == 0) {
        say_if(gu->verbose, "No new updates available\n");
        unlink(geoip_gz_filename);
        free(geoip_filename);
        free(geoip_gz_filename);
        return GU_OK;
    }

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
    char *geoip_filename = NULL, *geoip_gz_filename = NULL, *url = NULL;
    char hex_digest[33] = { 0 };

    xasprintf(&geoip_filename, "%s/GeoIP.dat", gu->database_dir);
    xasprintf(&geoip_gz_filename, "%s/GeoIP.dat.gz", gu->database_dir);

    // Calculate the MD5 hash of the database we currently have, if any. We get
    // back a zero MD5 hash if we don't have it yet.
    md5hex(geoip_filename, hex_digest);
    say_if(gu->verbose, "md5hex_digest: %s\n", hex_digest);

    xasprintf(&url,
              "%s://%s/app/update?license_key=%s&md5=%s",
              gu->proto, gu->host, &gu->license.license_key[0], hex_digest);

    char expected_file_md5[33] = { 0 };
    download_to_file(gu, url, geoip_gz_filename, expected_file_md5);
    free(url);

    // Was there actually an update? We can tell because if not we will have
    // the same MD5 reported back. Note in the past we would check the response
    // body which does still say whether we have an update.
    if (strcmp(hex_digest, expected_file_md5) == 0) {
        say_if(gu->verbose, "No new updates available\n");
        unlink(geoip_gz_filename);
        free(geoip_filename);
        free(geoip_gz_filename);
        return GU_OK;
    }

    int rc = gunzip_and_replace(gu, geoip_gz_filename, geoip_filename,
                                expected_file_md5);

    free(geoip_gz_filename);
    free(geoip_filename);

    return rc;
}

// Decompress the compressed database and move it into place in the database
// directory.
//
// We are given the path to the compressed (gzip'd) new database, and the path
// to where it should end up once decompressed. We are also given the MD5 hash
// it should have once decompressed for verification purposes.
//
// We verify the file is actually a gzip file. If it isn't we abort with an
// error, and remove the file.
//
// We also remove the gzip file once we successfully decompress and move the
// new database into place.
static int gunzip_and_replace(geoipupdate_s const * const gu,
                              char const * const gzipfile,
                              char const * const geoip_filename,
                              char const * const expected_file_md5)
{
    if (gu == NULL ||
        gu->database_dir == NULL || strlen(gu->database_dir) == 0 ||
        gzipfile == NULL || strlen(gzipfile) == 0 ||
        geoip_filename == NULL || strlen(geoip_filename) == 0 ||
        expected_file_md5 == NULL || strlen(expected_file_md5) == 0) {
        fprintf(stderr, "gunzip_and_replace: %s\n", strerror(EINVAL));
        return GU_ERROR;
    }

    if (!is_valid_gzip_file(gzipfile)) {
        // We should have already reported an error.
        unlink(gzipfile);
        return GU_ERROR;
    }

    // Decompress to the filename with the suffix ".test".
    char *file_path_test = NULL;
    xasprintf(&file_path_test, "%s.test", geoip_filename);
    say_if(gu->verbose, "Uncompress file %s to %s\n", gzipfile, file_path_test);

    gzFile gz_fh = gzopen(gzipfile, "rb");
    exit_if(gz_fh == NULL, "Can't open %s\n", gzipfile);

    FILE *fhw = fopen(file_path_test, "wb");
    exit_if(fhw == NULL, "Can't open %s\n", file_path_test);

    size_t const bsize = 8192;
    char * const buffer = calloc(bsize, sizeof(char));
    if (!buffer) {
        fprintf(stderr, "gunzip_and_replace: %s\n", strerror(errno));
        free(file_path_test);
        gzclose(gz_fh);
        fclose(fhw);
        return GU_ERROR;
    }

    for (;; ) {
        int amt = gzread(gz_fh, buffer, bsize);
        if (amt == 0) {
            // EOF
            break;
        }
        exit_if(amt == -1, "Gzip read error while reading from %s\n", gzipfile);
        exit_unless(fwrite(buffer, 1, amt, fhw) == (size_t)amt,
                    "Gzip write error\n");
    }
    exit_if(-1 == fclose(fhw), "Error closing stream: %s", strerror(errno));
    exit_if(gzclose(gz_fh) != Z_OK, "Gzip read error while closing from %s\n",
            gzipfile);
    free(buffer);

    char actual_md5[33] = { 0 };
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
    return GU_OK;
}
