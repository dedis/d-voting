#ifndef UK_SC_READ_BALLOT
#define UK_SC_READ_BALLOT

typedef void (*read_ballot_cb)(unsigned char *, unsigned char *, unsigned int, void *);
typedef void (*read_ballots_cb)(const char *, void *);

extern void read_ballot(unsigned char *output, const char *filepath, const unsigned int numNodes,
                        const unsigned int numChunks, read_ballot_cb f, void *f_data);

extern void read_ballots(const char *folder, const char *prefix, read_ballots_cb f,
                         void *f_data);

#endif