
#include "geoipupdate.h"
#include <string.h>
#include <stdlib.h>

geoipupdate_s *geoipupdate_s_new(void)
{
    size_t size = sizeof(geoipupdate_s);
    geoipupdate_s *gu = xmalloc(size);
    memset(gu, 0, size);
    gu->license_file = strdup(SYSCONFDIR"/GeoIP.conf");
    gu->database_dir = strdup(DATADIR);
    gu->proto = strdup("https");
    gu->host = strdup("updates.maxmind.com");
    gu->proxy_port = strdup("");
    gu->proxy_user_password = strdup("");
    gu->verbose = 0;
    gu->license.user_id = NO_USER_ID;
    gu->license.license_key[12] = 0;
    return gu;
}

static void xfree(void *p)
{
    if (p) {
        free(p);
    }
}

void geoipupdate_s_delete(geoipupdate_s * gu)
{
    if (gu) {
        product_delete_all(gu);
        xfree(gu->license_file);
        xfree(gu->database_dir);
        xfree(gu->proto);
        xfree(gu->proxy_port);
        xfree(gu->proxy_user_password);
        xfree(gu->host);
        free(gu);
    }
}

// return false on error
int geoipupdate_s_init(geoipupdate_s * gu)
{
    return 1;
}

void geoipupdate_s_cleanup(geoipupdate_s * gu)
{
}
