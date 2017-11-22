
#include "geoipupdate.h"
#include <errno.h>
#include <stdlib.h>
#include <string.h>

geoipupdate_s *geoipupdate_s_new(void) {
    size_t size = sizeof(geoipupdate_s);
    geoipupdate_s *gu = xcalloc(1, size);

    gu->license_file = strdup(SYSCONFDIR "/GeoIP.conf");
    exit_if(NULL == gu->license_file,
            "Unable to allocate memory for license file path: %s\n",
            strerror(errno));

    gu->database_dir = strdup(DATADIR);
    exit_if(NULL == gu->database_dir,
            "Unable to allocate memory for database directory path: %s\n",
            strerror(errno));

    gu->proto = strdup("https");
    exit_if(NULL == gu->proto,
            "Unable to allocate memory for request protocol: %s\n",
            strerror(errno));

    gu->host = strdup("updates.maxmind.com");
    exit_if(NULL == gu->host,
            "Unable to allocate memory for update host: %s\n",
            strerror(errno));

    gu->proxy = strdup("");
    exit_if(NULL == gu->proxy,
            "Unable to allocate memory for proxy host: %s\n",
            strerror(errno));

    gu->proxy_user_password = strdup("");
    exit_if(NULL == gu->proxy_user_password,
            "Unable to allocate memory for proxy credentials: %s\n",
            strerror(errno));

    gu->lock_file = strdup("");
    exit_if(NULL == gu->lock_file,
            "Unable to allocate memory for lock file path: %s\n",
            strerror(errno));

    gu->verbose = 0;
    gu->license.account_id = NO_ACCOUNT_ID;
    gu->license.license_key[12] = 0;

    // curl_easy_init() provides no useful error messages or codes:
    // https://curl.haxx.se/mail/lib-2009-11/0243.html
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
