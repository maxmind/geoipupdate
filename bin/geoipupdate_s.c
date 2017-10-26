
#include "geoipupdate.h"
#include <stdlib.h>
#include <string.h>

geoipupdate_s *geoipupdate_s_new(void) {
    size_t size = sizeof(geoipupdate_s);
    geoipupdate_s *gu = xmalloc(size);
    memset(gu, 0, size);

    gu->license_file = strdup(SYSCONFDIR "/GeoIP.conf");
    exit_if(NULL == gu->license_file,
            "Unable to allocate memory for license file path.\n");

    gu->database_dir = strdup(DATADIR);
    exit_if(NULL == gu->database_dir,
            "Unable to allocate memory for database directory path.\n");

    gu->proto = strdup("https");
    exit_if(NULL == gu->proto,
            "Unable to allocate memory for request protocol.\n");

    gu->host = strdup("updates.maxmind.com");
    exit_if(NULL == gu->host, "Unable to allocate memory for update host.\n");

    gu->proxy = strdup("");
    exit_if(NULL == gu->proxy, "Unable to allocate memory for proxy host.\n");

    gu->proxy_user_password = strdup("");
    exit_if(NULL == gu->proxy_user_password,
            "Unable to allocate memory for proxy credentials.\n");

    gu->lock_file = strdup("");
    exit_if(NULL == gu->lock_file,
            "Unable to allocate memory for lock file path.\n");

    gu->verbose = 0;
    gu->license.account_id = NO_ACCOUNT_ID;
    gu->license.license_key[12] = 0;

    gu->curl = curl_easy_init();
    exit_if(NULL == gu->curl, "Unable to initialize curl.\n");

    return gu;
}

void geoipupdate_s_delete(geoipupdate_s *gu) {
    if (gu) {
        edition_delete_all(gu);
        free(gu->license_file);
        free(gu->database_dir);
        free(gu->proto);
        free(gu->proxy);
        free(gu->proxy_user_password);
        free(gu->lock_file);
        free(gu->host);
        if (gu->curl != NULL) {
            curl_easy_cleanup(gu->curl);
        }
        free(gu);
    }
}
