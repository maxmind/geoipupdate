/* MD5.H - header file for MD5C.C
 */

/* Copyright (C) 1991-2, RSA Data Security, Inc. Created 1991. All
rights reserved.

License to copy and use this software is granted provided that it
is identified as the "RSA Data Security, Inc. MD5 Message-Digest
Algorithm" in all material mentioning or referencing this software
or this function.

License is also granted to make and use derivative works provided
that such works are identified as "derived from the RSA Data
Security, Inc. MD5 Message-Digest Algorithm" in all material
mentioning or referencing the derived work.

RSA Data Security, Inc. makes no representations concerning either
the merchantability of this software or the suitability of this
software for any particular purpose. It is provided "as is"
without express or implied warranty of any kind.

These notices must be retained in any copies of any part of this
documentation and/or software.
 */

/* MD5 context. */

#include "types.h"

typedef struct {
  u32 A,B,C,D;          /* chaining variables */
  u32  nblocks;
  byte buf[64];
  int  count;
} MD5_CONTEXT;

void md5_init( MD5_CONTEXT *ctx );
void md5_write( MD5_CONTEXT *hd, byte *inbuf, size_t inlen);
void md5_final( MD5_CONTEXT *hd );

