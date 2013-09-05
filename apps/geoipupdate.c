
#include "geoipupdate.h"
#include "geoipupdate_s.h"

#include <curl/curl.h>

#include <unistd.h>
#include <stdio.h>
#include <stdlib.h>
#ifdef HAVE_GETOPT_H
#include <getopt.h>
#endif
#include <ctype.h>

void *xmalloc(size_t size)
{
    void *ptr = malloc(size);
    if (!ptr) {
        fprintf(stderr, "Out of memory\n");
        exit(1);
    }
    return ptr;
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
