
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
int parse_license_file(geoipupdate_s * up);

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

void *xmalloc(size_t size)
{
    void *ptr = malloc(size);
    exit_unless(!!ptr, "Out of memory\n");
    return ptr;
}

int main(int argc, const char *argv[])
{
    curl_global_init(CURL_GLOBAL_DEFAULT);
    geoipupdate_s *gu = geoipupdate_s_new();
    if (gu) {

        if (geoipupdate_s_init(gu)) {
            //  parse_opts(argc, argv, gu);
            parse_license_file(gu);

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
        if (sscanf(strt, "UserId %d", &up->license.user_id) == 1)
            continue;
        if (sscanf(strt, "LicenseKey %[12]s", &up->license.license_key[0]) == 1)
            continue;

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
}
