
#include "geoipupdate.h"
#include <curl/curl.h>

#include <unistd.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#ifdef HAVE_GETOPT_H
#include <getopt.h>
#endif
#include <ctype.h>
#include <stdarg.h>
#include "md5.h"
#include <zlib.h>

int parse_license_file(geoipupdate_s * up);
void update_country_database(geoipupdate_s * gu);
static void get_to_disc(geoipupdate_s * gu, const char *url, const char *fname);
static void gunzip_and_replace(geoipupdate_s * gu, const char *gzipfile,
                        const char *geoip_filename);

void exit_unless(int expr, const char *fmt, ...)
{
    va_list ap;
    if (expr)
        return;
    va_start(ap, fmt);
    vfprintf(stderr, fmt, ap);
    va_end(ap);
    exit(1);
}

void say_if(int expr, const char *fmt, ...)
{
    va_list ap;
    if (!expr)
        return;
    va_start(ap, fmt);
    vfprintf(stdout, fmt, ap);
    va_end(ap);
}

void say(const char *fmt, ...)
{
    va_list ap;
    va_start(ap, fmt);
    vfprintf(stdout, fmt, ap);
    va_end(ap);
}

void *xmalloc(size_t size)
{
    void *ptr = malloc(size);
    exit_unless(!!ptr, "Out of memory\n");
    return ptr;
}

void *xrealloc(void *ptr, size_t size)
{
    void *mem = realloc(ptr, size);
    exit_unless(mem != NULL, "Out of memory\n");
    return mem;
}

int main(int argc, const char *argv[])
{
    curl_global_init(CURL_GLOBAL_DEFAULT);
    geoipupdate_s *gu = geoipupdate_s_new();
    if (gu) {

        if (geoipupdate_s_init(gu)) {
            //  parse_opts(argc, argv, gu);
            if (parse_license_file(gu)) {
//                if (gu->license.user_id == NO_USER_ID)
                update_country_database(gu);
            }

            geoipupdate_s_cleanup(gu);
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

    const char *sep = " \t";
    size_t bsize = 1024;
    char *buffer = xmalloc(bsize);
    ssize_t read_bytes;
    while ((read_bytes = getline(&buffer, &bsize, fh)) != -1) {
        size_t idx = strspn(buffer, sep);
        char *strt = &buffer[idx];
        if (*strt == '#')
            continue;
        if (sscanf(strt, "UserId %d", &up->license.user_id) == 1) {
            say_if(up->verbose, "UserId %d\n", up->license.user_id);
            continue;
        }
        if (sscanf(strt, "LicenseKey %12s", &up->license.license_key[0]) == 1) {
            say_if(up->verbose, "LicenseKey %s\n", up->license.license_key);
            continue;
        }

        char *p, *last;
        if ((p = strtok_r(strt, sep, &last))) {
            if (strcmp(p, "ProductIds") != 0)
                continue;
            while ((p = strtok_r(NULL, sep, &last))) {
                product_insert_once(up, p);
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
    const char zero_hex_digest[33] = "00000000000000000000000000000000\0";
    size_t len;
    MD5_CONTEXT context;

    FILE *fh = fopen(fname, "rb");
    if (fh == NULL) {
        strcpy(hex_digest, zero_hex_digest);
        return 0;
    }
    md5_init(&context);
    while ((len = fread(buffer, 1, bsize, fh)) > 0)
        md5_write(&context, buffer, len);
    md5_final(&context);
    memcpy(digest, context.buf, 16);
    fclose(fh);
    for (int i = 0; i < 16; i++)
        snprintf(&hex_digest[2 * i], 3, "%02x", digest[i]);
    return 1;
}

void get_to_disc(geoipupdate_s * gu, const char *url, const char *fname)
{
    FILE *f = fopen(fname, "w+");
    exit_unless(f != NULL, "Can't open %s\n", fname);
    CURL *curl = curl_easy_init();
    curl_easy_setopt(curl, CURLOPT_USERAGENT, GEOIP_USERAGENT);
    curl_easy_setopt(curl, CURLOPT_WRITEFUNCTION, NULL);
    curl_easy_setopt(curl, CURLOPT_WRITEDATA, (void *)f);
    curl_easy_setopt(curl, CURLOPT_URL, url);
    int res = curl_easy_perform(curl);

    exit_unless(res == CURLE_OK, "curl_easy_perform() failed: %s\n",
                curl_easy_strerror(res));

    curl_easy_cleanup(curl);
    fclose(f);
}

void update_country_database(geoipupdate_s * gu)
{
    char *geoip_filename, *geoip_gz_filename, *url;
    char hex_digest[33];
    asprintf(&geoip_filename, "%s/GeoIP.dat", gu->database_dir);
    exit_unless(geoip_filename != NULL, "Out of memory\n");
    asprintf(&geoip_gz_filename, "%s/GeoIP.dat.gz", gu->database_dir);
    exit_unless(geoip_gz_filename != NULL, "Out of memory\n");

    md5hex(geoip_filename, hex_digest);
    say_if(gu->verbose, "md5hex_digest: %s\n", hex_digest);
    asprintf(&url,
             "%s://%s/app/update?license_key=%s&md5=%s",
             gu->proto, gu->host, &gu->license.license_key[0], hex_digest);
    exit_unless(url != NULL, "Out of memory\n");

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
    char *buffer = xmalloc(bsize);
    ssize_t read_bytes = getline(&buffer, &bsize, fh);
    fclose(fh);
    exit_unless(read_bytes >= 0, "Read error %s\n", gzipfile);
    const char *no_new_upd = "No new updates available";
    if (strncmp(no_new_upd, buffer, strlen(no_new_upd)) == 0) {
        say_if(gu->verbose, "%s\n", no_new_upd);
        free(buffer);
        return;
    }
    char *file_path_test;
    asprintf(&file_path_test, "%s.test", geoip_filename);
    say_if(gu->verbose, "Uncompress file %s to %s\n", gzipfile, file_path_test);
    gz_fh = gzopen(gzipfile, "rb");
    exit_unless(gz_fh != NULL, "Can't open %s\n", gzipfile);
    FILE *fhw = fopen(file_path_test, "wb");
    exit_unless(fhw >= 0, "Can't open %s\n", file_path_test);

    for (;;) {
        int amt;
        amt = gzread(gz_fh, buffer, bsize);
        if (amt == 0)
            break;              // EOF
       exit_unless(amt != -1, "Gzip read error while reading from %s\n",
                    gzipfile);
         exit_unless(fwrite(buffer, 1, amt, fhw) == amt, "Gzip write error\n");
    }
    fclose(fhw);
    gzclose(gz_fh);
    free(buffer);
    int err = rename(file_path_test, geoip_filename);
    exit_unless(!err, "Rename %s to %s failed\n", file_path_test,
                geoip_filename);
    unlink(gzipfile);
    free(file_path_test);
}

