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

This script will setup the nodes and services. If you restart do not forget to
remove the old state:

```
rm -rf /tmp/node*
```

# Testing
## Automate the memcoin setup using tmux
If you have `tmux` installed, you can start a `tmux` session that will execute the above setup by running:
```
./test.sh
```

## Run the scenario test

If nodes are running and `setup.sh` has been called, you can run a test
scenario:

```
LLVL=info memcoin --config /tmp/node1 e-voting scenarioTest
```

# Use the frontend

See README in `web/`.
