
#include "geoipupdate.h"
#include <string.h>
#include <stdlib.h>

int product_count(geoipupdate_s * gu)
{
    int cnt = 0;
    for (product_s * p = gu->license.first; p; p = p->next) {
        cnt++;
    }
    return cnt;
}

void product_delete_all(geoipupdate_s * gu)
{
    product_s *next, *current;

    for (next = gu->license.first; (current = next); ) {
        next = current->next;
        product_delete(current);
    }
}

void product_insert_once(geoipupdate_s * gu, const char *product_id)
{
    product_s **next = &gu->license.first;
    for (; *next; next = &(*next)->next) {
        if (strcmp((*next)->product_id, product_id) == 0) {
            return;
        }
    }
    *next = product_new(product_id);
    say_if(gu->verbose, "Insert product_id %s\n", product_id);

}

product_s *product_new(const char *product_id)
{
    product_s *p = xmalloc(sizeof(product_s));
    p->product_id = strdup(product_id);
    p->next = NULL;
    return p;
}

void product_delete(product_s * p)
{
    if (p) {
        free(p->product_id);
    }
    free(p);
}
