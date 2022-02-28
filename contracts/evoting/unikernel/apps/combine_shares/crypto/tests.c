/*
This file implements unit tests of crypto.c functions. Baseline values are
computed with the kyber library.
Run with `./run_tests.sh`.
*/

#include <stdio.h>
#include <assert.h>
#include <sodium.h>
#include "crypto.h"
#include "crypto_tests.h"
#include "read_ballots.h"

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

void read_ballot_test_callback(unsigned char *out, unsigned char *in, int n, void *f_data)
{
    assert(n == 3);

    int *i = (int *)f_data;

    char p1_hex[] = "a2c743a04a1f8318059f876365df6ce0f30d2d65cce14c4bb923fa1c3a67bddd";
    char p2_hex[] = "defe4bcd17f0c9933139ee1d60c5ae5446043fee8ba29a726ebbead6d8250153";
    char p3_hex[] = "047607dc8a3f924d000f9146dbda989614137a49ccaf4b3082e9ff4c9fa8d1a0";
    char p4_hex[] = "251858365cca8e0cfd534e974fc5c1d48dcbb8149ab7a8ec673f1112cc3cc23a";
    char p5_hex[] = "8025b9aa3acd3a88173e1d2e134a18824d8600a31c689288da56af530722fccf";
    char p6_hex[] = "e09875368cd7c965548a269963c8cd3157ce3103297252661476adf23000ff7d";

    unsigned char p1[crypto_core_ed25519_BYTES] = {0};
    unsigned char p2[crypto_core_ed25519_BYTES] = {0};
    unsigned char p3[crypto_core_ed25519_BYTES] = {0};
    unsigned char p4[crypto_core_ed25519_BYTES] = {0};
    unsigned char p5[crypto_core_ed25519_BYTES] = {0};
    unsigned char p6[crypto_core_ed25519_BYTES] = {0};

    const size_t psize = crypto_core_ed25519_BYTES;

    sodium_hex2bin(p1, psize, p1_hex, psize * 2, NULL, NULL, NULL);
    sodium_hex2bin(p2, psize, p2_hex, psize * 2, NULL, NULL, NULL);
    sodium_hex2bin(p3, psize, p3_hex, psize * 2, NULL, NULL, NULL);
    sodium_hex2bin(p4, psize, p4_hex, psize * 2, NULL, NULL, NULL);
    sodium_hex2bin(p5, psize, p5_hex, psize * 2, NULL, NULL, NULL);
    sodium_hex2bin(p6, psize, p6_hex, psize * 2, NULL, NULL, NULL);

    switch (*i)
    {
    case 0:
        assert(compare_point(p1, &in[psize * 0]) == 0);
        assert(compare_point(p2, &in[psize * 1]) == 0);
        assert(compare_point(p3, &in[psize * 2]) == 0);

        out[0] = 10;

        break;
    case 1:
        assert(compare_point(p4, &in[psize * 0]) == 0);
        assert(compare_point(p5, &in[psize * 1]) == 0);
        assert(compare_point(p6, &in[psize * 2]) == 0);

        out[0] = 20;

        break;
    default:
        printf("FAILED: read_ballot_simple: unexpected index %i\n", *i);
        break;
    }

    printf("OK: read_ballot_simple (%i)\n", *i);

    (*(int *)f_data) = *i + 1;
}

void test_read_ballot_simple()
{
    // vote_1.bin contains the following points (spaces are meaningless):
    //
    //   p1 p2 p3    p4 p5 p6
    //
    // p1: a2c743a04a1f8318059f876365df6ce0f30d2d65cce14c4bb923fa1c3a67bddd
    // p2: defe4bcd17f0c9933139ee1d60c5ae5446043fee8ba29a726ebbead6d8250153
    // p3: 047607dc8a3f924d000f9146dbda989614137a49ccaf4b3082e9ff4c9fa8d1a0
    // p4: 251858365cca8e0cfd534e974fc5c1d48dcbb8149ab7a8ec673f1112cc3cc23a
    // p5: 8025b9aa3acd3a88173e1d2e134a18824d8600a31c689288da56af530722fccf
    // p6: e09875368cd7c965548a269963c8cd3157ce3103297252661476adf23000ff7d
    //
    // Expected decrypted chunks are:
    //
    // c1: e8a3f31322a30be2f0eed3cb887a39bb3e47a2188b20b9eefe2cb72d9753a45b
    // c2: 59e42bc01de0a6a42a336c317b58bf9ef7d61bd6553462992798ce8707302899

    int counter = 0;

    const int numChunks = 2;
    unsigned char output[32 * numChunks];

    read_ballot(output, "vote_1.bin", 3, 2, read_ballot_test_callback, &counter);

    assert(output[0] == 10);
    assert(output[32] == 20);

    printf("OK: read_ballot_simple\n");
}

// this function wraps the recover_commit function to be used as callback to the
// read_ballot function.
void read_ballot_real_callback(unsigned char *out, unsigned char *in, int n, void *f_data)
{
    recover_commit(out, in, n);
}

void test_read_ballot_real()
{
    char expected_1_hex[] = "e8a3f31322a30be2f0eed3cb887a39bb3e47a2188b20b9eefe2cb72d9753a45b";
    char expected_2_hex[] = "59e42bc01de0a6a42a336c317b58bf9ef7d61bd6553462992798ce8707302899";

    unsigned char expected_p1[crypto_core_ed25519_BYTES] = {0};
    unsigned char expected_p2[crypto_core_ed25519_BYTES] = {0};

    const size_t psize = crypto_core_ed25519_BYTES;

    sodium_hex2bin(expected_p1, psize, expected_1_hex, psize * 2, NULL, NULL, NULL);
    sodium_hex2bin(expected_p2, psize, expected_2_hex, psize * 2, NULL, NULL, NULL);

    const int numChunks = 2;
    unsigned char output[32 * numChunks];

    read_ballot(output, "vote_1.bin", 3, 2, read_ballot_real_callback, NULL);

    if (compare_scalar(expected_p1, output) != 0)
    {
        printf("FAILED: test_read_ballot_real\n");
    }
    else
    {
        printf("OK: read_ballot_real\n");
    }

    if (compare_scalar(expected_p2, &output[32]) != 0)
    {
        printf("FAILED: test_read_ballot_real\n");
    }
    else
    {
        printf("OK: read_ballot_real\n");
    }
}

void read_ballots_test_ballback(const char *filepath)
{
    printf("res: %s\n", filepath);
}

void test_read_ballots()
{
    read_ballots("./ballots", "ballot", read_ballots_test_ballback);
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

    test_read_ballot_simple();
    test_read_ballot_real();
    test_read_ballots();

    printf("------\n");
}