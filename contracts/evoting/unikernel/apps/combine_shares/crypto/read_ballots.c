#include "read_ballots.h"
#include <stdio.h>
#include <dirent.h>
#include <string.h>

// This function reads a file containing all the public shares of a ballot, and
// calls a callback function to decrypt each chunk.
//
// The file looks like as follow, spaces are for display purpose only. 'c'
// stands for 'chunk'  and 'n' stands for 'node'. The file contains the public
// shares of each chunk, ordered by the DKG node indexes (n2, n1, n3 in this
// case):
//
//     c1n2 c1n1 c1n3   c2n2 c2n1 c2n3
//
// In this example the ballot has 2 chunks and there are 3 nodes. The total file
// size is `numNodes * numChunks * 32`.
//
// The function saves the result to the output. The output must be of size
// `32 * numChunks`.
//
void read_ballot(unsigned char *output, const char *filepath, const unsigned int numNodes,
                 const unsigned int numChunks, read_ballot_cb f, void *f_data)
{
    FILE *fp;
    unsigned char buff[32 * numNodes];

    fp = fopen(filepath, "rt");
    if (fp == NULL)
    {
        printf("Error opening file %s\n", filepath);
        return;
    }

    for (unsigned int i = 0; i < numChunks; i++)
    {
        fread(buff, sizeof(char), 32 * numNodes, (FILE *)fp);

        f(&output[i * 32], buff, numNodes, f_data);
    }

    fclose(fp);
}

// this function reads all files from the directory matching the prefix and
// calls the provided callback. The callback has the full filepath and the
// f_data that can be used to stored data across callback calls.
void read_ballots(const char *folder, const char *prefix, read_ballots_cb f, void *f_data)
{
    DIR *d;
    struct dirent *dir;
    d = opendir(folder);
    if (d)
    {
        while ((dir = readdir(d)) != NULL)
        {
            if (strncmp(prefix, dir->d_name, strlen(prefix)) == 0)
            {
                char str[256];
                strcpy(str, folder);
                strcat(str, dir->d_name);
                f(str, f_data);
            }
        }
        closedir(d);
    }
}