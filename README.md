[![Go Test](https://github.com/dedis/d-voting/actions/workflows/go_test.yml/badge.svg)](https://github.com/dedis/d-voting/actions/workflows/go_test.yml)
[![Go Test](https://github.com/dedis/d-voting/actions/workflows/go_integration_tests.yml/badge.svg)](https://github.com/dedis/d-voting/actions/workflows/go_integration_tests.yml)
[![Go Test](https://github.com/dedis/d-voting/actions/workflows/go_memcoin_test.yml/badge.svg)](https://github.com/dedis/d-voting/actions/workflows/go_memcoin_test.yml)
[![Coverage Status](https://coveralls.io/repos/github/dedis/d-voting/badge.svg)](https://coveralls.io/github/dedis/d-voting)

# Setup

First be sure to have Go installed (at least 1.17).

Be sure to have the `crypto` utility from Dela:

```sh
git clone https://github.com/dedis/dela.git
cd dela/cli/crypto
go install
```

Go will install the binaries in `$GOPATH/bin`, so be sure this it is correctly
added to you path (like with `export PATH=$PATH:/Users/david/go/bin`).

Create a private key (in the d-voting root folder):

```sh
crypto bls signer new --save private.key
```

Copy the private key from the d-voting root folder to the `cli/memcoin` folder:

```sh
cp private.key cli/memcoin/
```

Install memcoin (this requires the private key in `cli/memcoin`):

```sh
cd cli/memcoin
go install
```

# Run the nodes

In three different terminal sessions, from the root folder:

```sh
LLVL=info memcoin --config /tmp/node1 start --port 2001

LLVL=info memcoin --config /tmp/node2 start --port 2002

LLVL=info memcoin --config /tmp/node3 start --port 2003
```

Then you should be able to run the setup script:

```sh
./setup.sh
```

This script will setup the nodes and services. If you restart do not forget to
remove the old state:

```sh
rm -rf /tmp/node{1,2,3}
```

# Testing

## Automate the previous setup using `tmux`

If you have `tmux` installed, you can start a `tmux` session that will
execute the above setup by running `./start_test.sh` in the project root.
Once the session is started, you can move around the panes with
`Ctrl+B` followed by arrow keys.

The top-left pane is for running commands, while the rest are for examining the node states.

To end the session, run `./kill_test.sh`,
which will kill each pane then the `tmux` session (which you can do manually with `Ctrl+D`),
then delete the node data (i.e. the files `/tmp/node{1,2,3}`).

## Run the scenario test

If nodes are running and `setup.sh` has been called, you can run a test
scenario:

```sh
LLVL=info memcoin --config /tmp/node1 e-voting scenarioTest
```

# Use the frontend

See README in `web/`.
