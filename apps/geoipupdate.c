
#include "geoipupdate.h"
#include <curl/curl.h>

#include <unistd.h>
#include <stdio.h>
#include <stdlib.h>
#ifdef HAVE_GETOPT_H
#include <getopt.h>
#endif
#include <ctype.h>

static void *xmalloc(size_t size)
{
    void *ptr = malloc(size);
    if (!ptr) {
        fprintf(stderr, "Out of memory\n");
        exit(1);
    }
    return ptr;
}

typedef struct {
    int user_id;
    char license_key[12];
    product_id_s *first;
} license_s;

typedef struct product_id_s {
    char *product_id;
    struct product_id_s *next;
} product_id_s;

typedef struct {
    license_s license;
    CURL *curl;

    // user might change these before geoipupdate_s_init
    int skip_peer_verification;
    int skip_hostname_verification;
} geoipupdate_s;

int geoipupdate_s_new(void)
{
    size_t size = sizeof(geoipupdate_s);
    geoipupdate_s *gu = xmalloc(size);
    memset(gu, 0, size);
    return gu;
}

void geoipupdate_s_delete(geoipupdate_s * gu)
{
    if (gu)
        free(gu);
}

// return false on error
int geoipupdate_s_init(geoipupdate_s * gu)
{
    if ((gu->curl = curl_easy_init())) {
        curl_easy_setopt(gu->curl, CURLOPT_SSL_VERIFYPEER,
                         gu->skip_peer_verification);
        curl_easy_setopt(gu->curl, CURLOPT_SSL_VERIFYHOST,
                         gu->skip_hostname_verification);
        return 1;
    }
    return 0;
}

void geoipupdate_s_cleanup(geoipupdate_s * gu)
{
    if (gu->curl) {
        curl_easy_cleanup(gu->curl);
    }
}

void geoipupdate_s_delete(geoipupdate_s * gu)
{
    if (gu->curl)
        curl_easy_cleanup(curl);
}

int main(int argc, const char *argv[])
{
    curl_global_init(CURL_GLOBAL_DEFAULT);
    geoipupdate_s *gu = geoipupdate_s_new();
    if (gu) {

        if (geoipupdate_s_init(gu)) {

            geoipupdate_s_cleanup(gu);
        }
        geoipupdate_s_delete(gu);
    }
    curl_global_cleanup();
    return 0;
}
