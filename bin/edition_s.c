
#include "geoipupdate.h"
#include <errno.h>
#include <stdlib.h>
#include <string.h>

int edition_count(geoipupdate_s *gu) {
    int cnt = 0;
    for (edition_s *p = gu->license.first; p; p = p->next) {
        cnt++;
    }
    return cnt;
}

void edition_delete_all(geoipupdate_s *gu) {
    edition_s *next, *current;

    for (next = gu->license.first; (current = next);) {
        next = current->next;
        edition_delete(current);
    }
}

void edition_insert_once(geoipupdate_s *gu, const char *edition_id) {
    edition_s **next = &gu->license.first;
    for (; *next; next = &(*next)->next) {
        if (strcmp((*next)->edition_id, edition_id) == 0) {
            return;
        }
    }
    *next = edition_new(edition_id);
    say_if(gu->verbose, "Insert edition_id %s\n", edition_id);
}

edition_s *edition_new(const char *edition_id) {
    edition_s *p = xcalloc(1, sizeof(edition_s));
    p->edition_id = strdup(edition_id);
    exit_if(NULL == p->edition_id,
            "Unable to allocate memory for edition ID: %s\n",
            strerror(errno));
    p->next = NULL;
    return p;
}

void edition_delete(edition_s *p) {
    if (p) {
        free(p->edition_id);
    }
    free(p);
}
