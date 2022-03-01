#ifndef UK_SC_CRYPTO
#define UK_SC_CRYPTO

#define point_size 32
#define scalar_size 32

extern void recover_commit(unsigned char *output, unsigned char *points, int n);

#endif