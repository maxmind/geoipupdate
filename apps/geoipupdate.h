
#ifndef GEOIPUPDATE_H
# define GEOIPUPDATE_H (1)

#include <stdlib.h>

typedef struct product_s {
    char *product_id;
    struct product_s *next;
} product_s;

typedef struct {
    int user_id;
    char license_key[13];
    product_s *first;
} license_s;

typedef struct {
    license_s license;

    // user might change these before geoipupdate_s_init
    int skip_peer_verification;
    int skip_hostname_verification;
    int do_not_overwrite_database_directory;
    char * license_file;
    char * database_dir;
    char * host;
    char * proto;
    char * proxy_port; // 1.2.3.4, 1.2.3.4:1234
    char * proxy_user_password; // user:pwd
    int verbose;

} geoipupdate_s;

geoipupdate_s * geoipupdate_s_new(void);
void geoipupdate_s_delete(geoipupdate_s * gu);
void product_delete_all(geoipupdate_s * gu);
int geoipupdate_s_init(geoipupdate_s * gu);
void geoipupdate_s_cleanup(geoipupdate_s * gu);

int product_count(geoipupdate_s * gu);
void product_insert_once(geoipupdate_s * gu, const char *product_id);
product_s *product_new(const char *product_id);
void product_delete(product_s * p);

void exit_unless(int expr, const char *fmt, ...);
void say_if(int expr, const char *fmt, ...);
void *xmalloc(size_t size);

# define NO_USER_ID (-1)
# define GEOIP_USERAGENT "geoipupdate/2.0"
#endif
