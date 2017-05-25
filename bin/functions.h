#ifndef _GEOIPUPDATE_FUNCTIONS_H
#define _GEOIPUPDATE_FUNCTIONS_H

#include <stdbool.h>

bool is_valid_gzip_file(char const * const);
char * slurp_file(char const * const);

#endif
