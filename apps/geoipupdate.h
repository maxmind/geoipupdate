
#ifndef GEOIPUPDATE_H
# define GEOIPUPDATE_H (1)

#include <curl/curl.h>

typedef struct product_s {
    char *product_id;
    struct product_s *next;
} product_s;

typedef struct {
    int user_id;
    char license_key[12];
    product_s *first;
} license_s;

typedef struct {
    license_s license;
    CURL *curl;

    // user might change these before geoipupdate_s_init
    int skip_peer_verification;
    int skip_hostname_verification;
} geoipupdate_s;

geoipupdate_s * geoipupdate_s_new(void);
void geoipupdate_s_delete(geoipupdate_s * gu);
int geoipupdate_s_init(geoipupdate_s * gu);
void geoipupdate_s_cleanup(geoipupdate_s * gu);

#endif
