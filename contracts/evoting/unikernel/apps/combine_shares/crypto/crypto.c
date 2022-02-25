/*
This file implements the reconstruction of public shares in the context of a DKG
protocol. The code maps closely the kyber implementation, but assumes the shares
are already sorted and cleaned.
*/

#include <sodium.h>
#include <stdio.h>

// prints a point as bytes
void print_point(const unsigned char *p)
{
    for (int i = 0; i < crypto_core_ed25519_BYTES; i++)
    {
        printf("%u ", p[i]);
    }
    printf("\n");
}

// prints a scalar as bytes
void print_scalar(const unsigned char *s)
{
    for (int i = 0; i < crypto_core_ed25519_SCALARBYTES; i++)
    {
        printf("%u ", s[i]);
    }
    printf("\n");
}

// sets a scalar to 0
void scalar_zero(unsigned char *s)
{
    for (int i = 0; i < crypto_core_ed25519_SCALARBYTES; i++)
    {
        s[i] = 0;
    }
}

// sets a scalar to 1
void scalar_one(unsigned char *s)
{
    scalar_zero(s);
    s[0] = 1;
}

// sets a scalar to the specified int
void scalar_int(int n, unsigned char *s)
{
    scalar_zero(s);
    s[3] = (n >> 24) & 0xFF;
    s[2] = (n >> 16) & 0xFF;
    s[1] = (n >> 8) & 0xFF;
    s[0] = n & 0xFF;
}

// perform x/y and stores the result to z
void scalar_divide(unsigned char *z, unsigned char *x, unsigned char *y)
{
    unsigned char inv[crypto_core_ed25519_SCALARBYTES] = {0};
    crypto_core_ed25519_scalar_invert(inv, y);
    crypto_core_ed25519_scalar_mul(z, x, inv);
}

// Gets as input a list of n Point, sorted by their DKG index and reconstructs
// the secret commitment using Lagrange interpolation. Result is saved to the
// output variable.
void recover_commit(unsigned char *output, unsigned char *points, int n)
{
    // set the output to the neutral element
    for (int i = 0; i < crypto_core_ed25519_BYTES; i++)
    {
        output[i] = 0;
    }
    output[0] = 1;

    unsigned char Tmp[crypto_core_ed25519_BYTES] = {0};

    unsigned char tmp[crypto_core_ed25519_SCALARBYTES] = {0};

    unsigned char num[crypto_core_ed25519_SCALARBYTES] = {0};
    unsigned char denum[crypto_core_ed25519_SCALARBYTES] = {0};

    // placeholders to contain the scalar indexes
    unsigned char indexI[crypto_core_ed25519_SCALARBYTES] = {0};
    unsigned char indexJ[crypto_core_ed25519_SCALARBYTES] = {0};

    for (int i = 0; i < n; i++)
    {
        scalar_one(num);
        scalar_one(denum);

        scalar_int(i + 1, indexI);

        for (int j = 0; j < n; j++)
        {
            if (i == j)
            {
                continue;
            }

            scalar_int(j + 1, indexJ);

            // stores x * y (mod L) into z
            crypto_core_ed25519_scalar_mul(num, indexJ, num);
            crypto_core_ed25519_scalar_sub(tmp, indexJ, indexI);
            crypto_core_ed25519_scalar_mul(denum, denum, tmp);
        }

        scalar_divide(num, num, denum);
        crypto_scalarmult_ed25519_noclamp(Tmp, num, &points[crypto_core_ed25519_BYTES * i]);
        crypto_core_ed25519_add(output, output, Tmp);
    }
}