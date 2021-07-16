# Setup

First be sure to have Go installed.

Install memcoin:

```
cd cli/memcoin
go install
```

Be sure to have the `crypto` utility from Dela:

```
git clone https://github.com/dedis/dela.git
cd dela/cli/crypto
go install
```

Go will install the binaries in `$GOPATH/bin`, so be sure this it is correctly
added to you path (like with `export PATH=$PATH:/Users/david/go/bin`).

Create a private key (from the root folder):

```
crypto bls signer new --save private.key
```

# Run the nodes

In three different terminal sessions, from the root folder:

```
LLVL=info memcoin --config /tmp/node1 start --port 2001

LLVL=info memcoin --config /tmp/node2 start --port 2002

LLVL=info memcoin --config /tmp/node3 start --port 2003
```

Then you should be able to run the setup script:

```
./setup.sh
```

This script will block once the http server is started.

In case you want to restart with a fresh state, do not forget to remove the
node's data:

```
rm -rf /tmp/node*
```

# Use the frontend

See README in `web/`.