#include "functions.h"

#include <errno.h>
#include <fcntl.h>
#include <limits.h>
#include <stdbool.h>
#include <stddef.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/types.h>
#include <sys/stat.h>
#include <unistd.h>

static ssize_t read_file(char const * const, void * const,
                         size_t const);

// Check whether the file looks like a valid gzip file.
//
// We open it and read in a small amount. We do this so we can check its header.
//
// Return true if it is.
bool is_valid_gzip_file(char const * const file)
{
    if (file == NULL || strlen(file) == 0) {
        fprintf(stderr, "is_valid_gzip_file: %s\n", strerror(EINVAL));
        return false;
    }

    size_t const bufsz = 2;

    uint8_t * const buf = calloc(bufsz, sizeof(uint8_t));
    if (buf == NULL) {
        fprintf(stderr, "is_valid_gzip_file: %s\n", strerror(errno));
        return false;
    }

    ssize_t const sz = read_file(file, buf, bufsz);
    if (sz == -1) {
        // We should have already reported an error.
        free(buf);
        return false;
    }

    if ((size_t const)sz != bufsz) {
        fprintf(stderr, "%s is not a valid gzip file (due to file size)\n",
                file);
        free(buf);
        return false;
    }

    if (buf[0] != 0x1f || buf[1] != 0x8b) {
        fprintf(stderr, "%s is not a valid gzip file\n", file);
        free(buf);
        return false;
    }

    free(buf);

    return true;
}

// Read in up to 8 KiB of a file.
//
// If you don't care how much of the file you read in, this function is easier
// to use than read_file().
//
// I chose 8 KiB arbitrarily.
//
// We guarantee the returned buffer will terminate with a NUL byte and be
// exactly 8 KiB. The useful portion may fall short of 8 KiB.
//
// The caller is responsible for the returned memory.
char * slurp_file(char const * const file)
{
    if (file == NULL || strlen(file) == 0) {
        fprintf(stderr, "slurp_file: %s\n", strerror(EINVAL));
        return NULL;
    }

    size_t sz = 8193;

    char * const buf = calloc(sz, sizeof(char));
    if (buf == NULL) {
        fprintf(stderr, "slurp_file: %s\n", strerror(errno));
        return NULL;
    }

    ssize_t const read_sz = read_file(file, buf, sz - 1);
    if (read_sz == -1) {
        // We should have reported an error.
        free(buf);
        return NULL;
    }

    return buf;
}

// Read in up to the first sz bytes of a file.
//
// Return how many bytes we read. -1 if there was an error.
//
// The buffer may or may not contain a string. It may be binary data.
static ssize_t read_file(char const * const file, void * const buf,
                         size_t const bufsz)
{
    if (file == NULL || strlen(file) == 0 || buf == NULL || bufsz == 0) {
        fprintf(stderr, "read_file: %s\n", strerror(EINVAL));
        return -1;
    }

    // Note previously we used fopen() and getline() to read, but getline() is
    // not appropriate when we have a binary file such as gzip. It reads until
    // it finds a newline and will resize the buffer if necessary. Use read(2)
    // instead.

    int const fd = open(file, O_RDONLY);
    if (fd == -1) {
        fprintf(stderr, "read_file: Can't open file: %s: %s\n",
                file, strerror(errno));
        return -1;
    }

    ssize_t total_read_bytes = 0;
    int retries_remaining = 3;

    while (1) {
        size_t const bytes_left_to_read = bufsz - total_read_bytes;
        if (bytes_left_to_read == 0) {
            break;
        }

        if (retries_remaining == 0) {
            fprintf(
                stderr,
                "read_file: Interrupted when reading from %s too many times\n",
                file);
            close(fd);
            return -1;
        }

        ssize_t const read_bytes = read(fd, buf + total_read_bytes,
                                        bytes_left_to_read);
        if (read_bytes < 0) {
            if (errno == EINTR) {
                retries_remaining--;
                continue;
            }
            fprintf(stderr, "read_file: Error reading from %s: %s\n", file,
                    strerror(errno));
            close(fd);
            return -1;
        }

        // EOF.
        if (read_bytes == 0) {
            break;
        }

        if (total_read_bytes > SSIZE_MAX - read_bytes) {
            fprintf(stderr,
                    "read_file: Overflow when counting number of read bytes\n");
            close(fd);
            return -1;
        }

        total_read_bytes += read_bytes;
    }

    if (close(fd) != 0) {
        fprintf(stderr, "read_file: Error closing file: %s: %s\n",
                file, strerror(errno));
        return -1;
    }

    return total_read_bytes;
}

#ifdef TEST_FUNCTIONS

#include <assert.h>

static void test_is_valid_gzip_file(void);
static void test_slurp_file(void);
static void test_read_file(void);
static char * get_temporary_filename(void);
static void write_file(char const * const, void const * const,
                       size_t const);

int main(void)
{
    test_is_valid_gzip_file();
    test_slurp_file();
    test_read_file();

    return 0;
}

static void test_is_valid_gzip_file(void)
{
    char * const filename = get_temporary_filename();
    assert(filename != NULL);

    // A buffer to work with.
    uint8_t buf[4] = { 0 };

    // Test: File does not exist.

    assert(!is_valid_gzip_file(filename));

    // Test: File is too short.

    memset(buf, 0, 4);
    buf[0] = 0x1f;
    write_file(filename, buf, 1);
    assert(!is_valid_gzip_file(filename));

    // Test: File is exactly long enough, but not a gzip file.

    memset(buf, 0, 4);
    buf[0] = 'a';
    buf[1] = 'b';
    write_file(filename, buf, 2);
    assert(!is_valid_gzip_file(filename));

    // Test: File is more than long enough, but not a gzip file.

    memset(buf, 0, 4);
    buf[0] = 'a';
    buf[1] = 'b';
    buf[3] = 'c';
    write_file(filename, buf, 3);
    assert(!is_valid_gzip_file(filename));

    // Test: File is exactly long enough, and a gzip file.

    memset(buf, 0, 4);
    buf[0] = 0x1f;
    buf[1] = 0x8b;
    write_file(filename, buf, 2);
    assert(is_valid_gzip_file(filename));

    // Test: File is more than long enough, and a gzip file (at least judging
    // by its header).

    memset(buf, 0, 4);
    buf[0] = 0x1f;
    buf[1] = 0x8b;
    buf[2] = 'a';
    write_file(filename, buf, 3);
    assert(is_valid_gzip_file(filename));

    // Clean up.
    unlink(filename);
    free(filename);
}

static void test_slurp_file(void)
{
    char * const filename = get_temporary_filename();
    assert(filename != NULL);

    // Test: File does not exist.

    char * const contents_0 = slurp_file(filename);
    assert(contents_0 == NULL);

    // Test: File is zero size.

    write_file(filename, "", 0);
    char * const contents_1 = slurp_file(filename);
    assert(contents_1 != NULL);
    assert(strlen(contents_1) == 0);
    free(contents_1);

    // Test: File has a short string.

    write_file(filename, "hello", strlen("hello"));
    char * const contents_2 = slurp_file(filename);
    assert(contents_2 != NULL);
    assert(strcmp(contents_2, "hello") == 0);
    free(contents_2);

    // Test: File is oversize.

    char contents[8194] = { 0 };
    memset(contents, 'a', 8193);

    write_file(filename, contents, strlen(contents));

    char expected[8193] = { 0 };
    memset(expected, 'a', 8192);

    char * const contents_3 = slurp_file(filename);
    assert(contents_3 != NULL);
    assert(strcmp(contents_3, expected) == 0);
    free(contents_3);

    // Clean up.
    assert(unlink(filename) == 0);
    free(filename);
}

static void test_read_file(void)
{
    char * const filename = get_temporary_filename();
    assert(filename != NULL);

    // Make a buffer to work with.
    size_t const bufsz = 32;
    char * const buf = calloc(bufsz, sizeof(char));
    assert(buf != NULL);

    // Test: The file does not exist.

    memset(buf, 0, bufsz);
    ssize_t const sz_0 = read_file(filename, buf, 2);
    assert(sz_0 == -1);

    // Test: The file is zero size.

    memset(buf, 0, bufsz);
    write_file(filename, "", 0);
    ssize_t const sz_1 = read_file(filename, buf, 2);
    assert(sz_1 == 0);

    // Test: The file is larger than we need.

    memset(buf, 0, bufsz);
    write_file(filename, "hello", strlen("hello"));
    ssize_t const sz_2 = read_file(filename, buf, 2);
    assert(sz_2 == 2);
    assert(buf[0] == 'h');
    assert(buf[1] == 'e');

    // Test: The file is exactly the size we need.

    memset(buf, 0, bufsz);
    write_file(filename, "hi", strlen("hi"));
    ssize_t const sz_3 = read_file(filename, buf, 2);
    assert(sz_3 == 2);
    assert(buf[0] == 'h');
    assert(buf[1] == 'i');

    // Test: The file has data, but not as much as we ask for.

    memset(buf, 0, bufsz);
    write_file(filename, "a", strlen("a"));
    ssize_t sz_4 = read_file(filename, buf, 2);
    assert(sz_4 == 1);
    assert(buf[0] == 'a');

    // Clean up.
    assert(unlink(filename) == 0);
    free(filename);
    free(buf);
}

static char * get_temporary_filename(void)
{
    size_t const sz = 64;

    char * const filename = calloc(sz, sizeof(char));
    assert(filename != NULL);

    strcat(filename, "/tmp/test-file-XXXXXX");
    int const fd = mkstemp(filename);
    assert(fd != -1);

    assert(close(fd) == 0);
    assert(unlink(filename) == 0);

    return filename;
}

static void write_file(char const * const path, void const * const contents,
                       size_t const sz)
{
    assert(path != NULL);
    assert(strlen(path) != 0);
    assert(contents != NULL);
    // Permit contents to be 0 length.

    int const fd = open(path, O_WRONLY | O_CREAT | O_TRUNC, S_IRUSR | S_IWUSR);
    if (fd == -1) {
        fprintf(stderr, "open() failed: %s: %s\n", path, strerror(errno));
        assert(0 == 1);
    }

    if (sz > 0) {
        ssize_t const write_sz = write(fd, contents, sz);
        assert(write_sz != -1);
        assert((size_t)write_sz == sz);
    }

    assert(close(fd) == 0);
}

#endif
