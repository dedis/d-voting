/*
This file implements the reconstruction of public shares in the context of a DKG
protocol. The code maps closely the kyber implementation, but assumes the shares
are already sorted and cleaned.
*/

#include <sodium.h>
#include <stdio.h>
#include "crypto.h"
#include <string.h>

// prints a point as bytes
void print_point(const unsigned char *p)
{
    for (int i = 0; i < point_size; i++)
    {
        printf("%u ", p[i]);
    }
    printf("\n");
}

// prints a scalar as bytes
void print_scalar(const unsigned char *s)
{
    for (int i = 0; i < scalar_size; i++)
    {
        printf("%u ", s[i]);
    }
    printf("\n");
}

// sets a scalar to 1
void scalar_one(unsigned char *s)
{
    memset(s, 0, scalar_size);
    s[0] = 1;
}

// sets a scalar to the specified int
void scalar_int(int n, unsigned char *s)
{
    memset(s, 0, scalar_size);
    s[3] = (n >> 24) & 0xFF;
    s[2] = (n >> 16) & 0xFF;
    s[1] = (n >> 8) & 0xFF;
    s[0] = n & 0xFF;
}

// perform x/y and stores the result to z
void scalar_divide(unsigned char *z, unsigned char *x, unsigned char *y)
{
    unsigned char inv[scalar_size];
    crypto_core_ed25519_scalar_invert(inv, y);
    crypto_core_ed25519_scalar_mul(z, x, inv);
}

// Gets as input a list of n Point, sorted by their DKG index and reconstructs
// the secret commitment using Lagrange interpolation. Result is saved to the
// output variable.
void recover_commit(unsigned char *output, unsigned char *points, int n)
{
    // set the output to the neutral element
    memset(output, 0, point_size);
    output[0] = 1;

    unsigned char Tmp[point_size];

    unsigned char tmp[scalar_size];

    unsigned char num[scalar_size];
    unsigned char denum[scalar_size];

    // placeholders to contain the scalar indexes
    unsigned char indexI[scalar_size];
    unsigned char indexJ[scalar_size];

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

        int res = crypto_scalarmult_ed25519_noclamp(Tmp, num, &points[point_size * i]);
        if (res != 0)
        {
            printf("ERROR: crypto_scalarmult_ed25519_noclamp failed: %d\n", res);
        }

        crypto_core_ed25519_add(output, output, Tmp);
    }
}