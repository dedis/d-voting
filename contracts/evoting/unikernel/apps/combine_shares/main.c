/*
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions
 * are met:
 *
 * 1. Redistributions of source code must retain the above copyright
 *    notice, this list of conditions and the following disclaimer.
 * 2. Redistributions in binary form must reproduce the above copyright
 *    notice, this list of conditions and the following disclaimer in the
 *    documentation and/or other materials provided with the distribution.
 * 3. Neither the name of the copyright holder nor the names of its
 *    contributors may be used to endorse or promote products derived from
 *    this software without specific prior written permission.
 *
 * THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
 * AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
 * IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
 * ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
 * LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
 * CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
 * SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
 * INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
 * CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
 * ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
 * POSSIBILITY OF SUCH DAMAGE.
 *
 * THIS HEADER MAY NOT BE EXTRACTED OR MODIFIED IN ANY WAY.
 */

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/socket.h>
#include <arpa/inet.h>
#include <unistd.h>
#include <errno.h>

#include <sodium.h>

#include "crypto/crypto.h"
#include "crypto/read_ballots.h"

#define DEFAULT_LISTEN_PORT 12345

/* "shortcut" for struct sockaddr structure */
#define SSA struct sockaddr

/* error printing macro */
#define ERR(call_description)             \
	do                                    \
	{                                     \
		fprintf(stderr, "(%s, %d): %s\n", \
				__FILE__, __LINE__,       \
				call_description);        \
	} while (0)

/* print error (call ERR) and exit */
#define DIE(assertion, call_description) \
	do                                   \
	{                                    \
		if (assertion)                   \
		{                                \
			ERR(call_description);       \
			return -1;                   \
		}                                \
	} while (0)

/*
 * Create a server socket.
 */

static int tcp_create_listener(unsigned short port, int backlog)
{
	struct sockaddr_in address;
	int listenfd;
	int sock_opt;
	int rc;

	listenfd = socket(PF_INET, SOCK_STREAM, 0);
	DIE(listenfd < 0, "socket");

	sock_opt = 1;
	rc = setsockopt(listenfd, SOL_SOCKET, SO_REUSEADDR,
					&sock_opt, sizeof(int));
	DIE(rc < 0, "setsockopt");

	memset(&address, 0, sizeof(address));
	address.sin_family = AF_INET;
	address.sin_port = htons(port);
	address.sin_addr.s_addr = INADDR_ANY;

	rc = bind(listenfd, (SSA *)&address, sizeof(address));
	DIE(rc < 0, "bind");

	rc = listen(listenfd, backlog);
	DIE(rc < 0, "listen");

	return listenfd;
}

/*
 * Use getpeername(2) to extract remote peer address. Fill buffer with
 * address format IP_address:port (e.g. 192.168.0.1:22).
 */

static int get_peer_address(int sockfd, char *buf, size_t __unused len)
{
	struct sockaddr_in addr;
	socklen_t addrlen = sizeof(struct sockaddr_in);

	if (getpeername(sockfd, (SSA *)&addr, &addrlen) < 0)
		return -1;

	sprintf(buf, "%s:%d", inet_ntoa(addr.sin_addr), ntohs(addr.sin_port));

	return 0;
}

/*
 *	ballot processing
 */

// this function wraps the recover_commit function to be used as callback to the
// read_ballot function.
void read_ballot_real_callback(unsigned char *out, unsigned char *in, unsigned int n, void *f_data)
{
	recover_commit(out, in, n);
}

struct RdBallotsCB
{
	unsigned int numNodes;
	unsigned int numChunks;
	char *outputFolder;
};

// implements a realistic callback for the read_ballots function. We store in
// f_data the number of nodes and number of chunks.
void read_ballots_real_callback(const char *filepath, void *f_data)
{
	struct RdBallotsCB *data = (struct RdBallotsCB *)f_data;

	unsigned char output[32 * data->numChunks];
	read_ballot(output, filepath, data->numNodes, data->numChunks,
				read_ballot_real_callback, NULL);

	FILE *fptr;

	char str[256];
	strcpy(str, filepath);
	strcat(str, ".decrypted");

	fptr = fopen(str, "w");

	if (fptr == NULL)
	{
		printf("ERROR: failed to create output file");
		return;
	}

	fwrite(output, sizeof(char), 32 * data->numChunks, fptr);

	fclose(fptr);
}

/*
	Entrypoint
*/

int main(int argc, char *argv[])
{
	int srv, client;
	int rc;
	unsigned short listen_port = DEFAULT_LISTEN_PORT;
	char addrname[128];
	char blockchain_input[256], unikernel_output[256];

	if (argc == 2)
	{
		printf("argument is %s\n", argv[1]);
		listen_port = atoi(argv[1]);
	}

	/* Initialize libsodium. */
	rc = sodium_init();
	if (rc < 0)
	{
		printf("Error initializing sodium.\n");
		return -1;
	}

	srv = tcp_create_listener(listen_port, 1);
	if (srv < 0)
		return -1;

	printf("Listening on port %hu...\n", listen_port);
	while (1)
	{
		ssize_t n;

		client = accept(srv, NULL, 0);
		DIE(client < 0, "accept");

		get_peer_address(client, addrname, 128);
		printf("Received connection from %s\n", addrname);

		/* Read command filename. */
		n = read(client, &blockchain_input, 256);
		if (n < 0)
		{
			printf("Error reading from socket.\n");
			goto close;
		}
		if (n == 0)
		{
			printf("Connection closed.\n");
			goto close;
		}

		if (n < 9)
		{
			printf("Expected at least 9 bytes: %d\n", n);
			goto close;
		}

		int numChunks = blockchain_input[0] +
						(blockchain_input[1] << 8) +
						(blockchain_input[2] << 16) +
						(blockchain_input[3] << 24);

		int numNodes = blockchain_input[4] +
					   (blockchain_input[5] << 8) +
					   (blockchain_input[6] << 16) +
					   (blockchain_input[7] << 24);

		struct RdBallotsCB data;
		data.numChunks = numChunks;
		data.numNodes = numNodes;

		read_ballots(&blockchain_input[8], "ballot", read_ballots_real_callback, &data);

		strcpy(unikernel_output, "Done!");

		n = write(client, &unikernel_output, strlen(unikernel_output));
		DIE(n < 0, "write");
		printf("Sent: %s\n", unikernel_output);

	close:
		/* Close connection */
		close(client);
	}

	return 0;
}
