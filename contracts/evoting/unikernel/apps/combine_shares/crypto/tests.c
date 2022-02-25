/*
This file implements unit tests of crypto.c functions. Baseline values are
computed with the kyber library.
Run with `./run_tests.sh`.
*/

#include <stdio.h>
#include <sodium.h>
#include "crypto.h"
#include "crypto_tests.h"

// print the hexadecimal representation of a buffer
void print_hex(const unsigned char *p, const int size)
{
    for (int i = 0; i < size; i++)
    {
        printf("%02x", p[i]);
    }
    printf("\n");
}

// outputs -1 if the two scalars are identical, otherwise 0
int compare_scalar(unsigned char *s1, unsigned char *s2)
{
    for (int i = 0; i < crypto_core_ed25519_SCALARBYTES; i++)
    {
        if (s1[i] != s2[i])
        {
            return -1;
        }
    }

    return 0;
}

// outputs -1 if the two points are identical, otherwise 0
int compare_point(unsigned char *s1, unsigned char *s2)
{
    for (int i = 0; i < crypto_core_ed25519_BYTES; i++)
    {
        if (s1[i] != s2[i])
        {
            return -1;
        }
    }

    return 0;
}

// tests the recover_commit function
void test_recover_commit()
{
    char p1_hex[] = "470dc332b7bfe7cbb5ab85ca6fb989b30a748cd2101f849126535e5cc2a7c8e2";
    char p2_hex[] = "db1b26bc20fb0e02bb9673f2f04c1e0c5afa1a7216fd1752a593592557223a89";
    char p3_hex[] = "37420f14fc9be1d9703988a694bfb5f0aad63edf37ec01912c5729f39dfca80d";

    char expected_hex[] = "92babdad0b09e3a4b86ddd1ef1c1ccdbbc35006768b0fecefdc96eb7dffddffb";

    unsigned char expected[crypto_core_ed25519_BYTES] = {0};

    unsigned char input[crypto_core_ed25519_BYTES * 3] = {0};

    const size_t psize = crypto_core_ed25519_BYTES;

    sodium_hex2bin(&input[psize * 0], psize, p1_hex, psize * 2, NULL, NULL, NULL);
    sodium_hex2bin(&input[psize * 1], psize, p2_hex, psize * 2, NULL, NULL, NULL);
    sodium_hex2bin(&input[psize * 2], psize, p3_hex, psize * 2, NULL, NULL, NULL);
    sodium_hex2bin(expected, psize, expected_hex, psize * 2, NULL, NULL, NULL);

    unsigned char output[crypto_core_ed25519_BYTES] = {0};

    recover_commit(output, input, 3);

    if (compare_point(expected, output) != 0)
    {
        printf("FAILED: test_recover_commit\n");
    }
    else
    {
        printf("OK: test_recover_commit\n");
    }
}

// tests the scalar_zero function
void test_scalar_zero()
{
    unsigned char s[crypto_core_ed25519_SCALARBYTES] = {1, 2, 3, 4, 5};
    unsigned char expected[crypto_core_ed25519_SCALARBYTES] = {0};

    scalar_zero(s);

    if (compare_scalar(expected, s) != 0)
    {
        printf("FAILED: test_scalar_zero\n");
    }
    else
    {
        printf("OK: test_scalar_zero\n");
    }
}

// tests the scalar_one function
void test_scalar_one()
{
    unsigned char s[crypto_core_ed25519_SCALARBYTES] = {1, 2, 3, 4, 5};
    unsigned char expected[crypto_core_ed25519_SCALARBYTES] = {1};

    scalar_one(s);

    if (compare_scalar(expected, s) != 0)
    {
        printf("FAILED: test_scalar_one\n");
    }
    else
    {
        printf("OK: test_scalar_one\n");
    }
}

// test the scalar_divide function
void test_scalar_divide()
{
    char s1_hex[] = "c63e655301fc5d26f27f088deecbccf844514b73f161311d5d18d07077c8b005";
    char s2_hex[] = "b404a42bd03849bf1986eaf33b022d2840f7550462e96364665b24eec629280a";

    char expected_hex[] = "eb1e0a0bd1e1ac34d7a9203255b26d314c64874c3eda33a5d02a0a3721a80200";

    unsigned char s1[crypto_core_ed25519_SCALARBYTES] = {0};
    unsigned char s2[crypto_core_ed25519_SCALARBYTES] = {0};
    unsigned char expected[crypto_core_ed25519_SCALARBYTES] = {0};

    unsigned char output[crypto_core_ed25519_SCALARBYTES] = {0};

    const size_t psize = crypto_core_ed25519_SCALARBYTES;

    sodium_hex2bin(s1, psize, s1_hex, psize * 2, NULL, NULL, NULL);
    sodium_hex2bin(s2, psize, s2_hex, psize * 2, NULL, NULL, NULL);
    sodium_hex2bin(expected, psize, expected_hex, psize * 2, NULL, NULL, NULL);

    scalar_divide(output, s1, s2);

    if (compare_scalar(expected, output) != 0)
    {
        printf("FAILED: test_scalar_divide\n");
    }
    else
    {
        printf("OK: test_scalar_divide\n");
    }
}

// tests the scalar_int function
void test_scalar_int()
{
    unsigned char s[crypto_core_ed25519_SCALARBYTES] = {1, 2, 3, 4, 5};
    unsigned char expected[crypto_core_ed25519_SCALARBYTES] = {44, 1};

    scalar_int(300, s);

    if (compare_scalar(expected, s) != 0)
    {
        printf("FAILED: test_scalar_int\n");
    }
    else
    {
        printf("OK: test_scalar_int\n");
    }
}

// entry point, launch the tests
int main(int argc, char *argv[])
{
    printf("Tests:\n------\n");
    test_recover_commit();
    test_scalar_zero();
    test_scalar_one();
    test_scalar_int();
    test_scalar_divide();
    printf("------\n");
}