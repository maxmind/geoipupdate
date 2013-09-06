
#include "geoipupdate.h"
#include <string.h>
#include <curl/curl.h>
#include <stdlib.h>

geoipupdate_s *geoipupdate_s_new(void)
{
    size_t size = sizeof(geoipupdate_s);
    geoipupdate_s *gu = xmalloc(size);
    memset(gu, 0, size);
    gu->license_file = "/usr/local/etc/GeoIP.conf";
    gu->verbose = 1;
    gu->license.license_key[13] = 0;
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
