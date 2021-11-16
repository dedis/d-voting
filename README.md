# Setup

First be sure to have Go installed.

Be sure to have the `crypto` utility from Dela:

```
git clone https://github.com/dedis/dela.git
cd dela/cli/crypto
go install
```

Go will install the binaries in `$GOPATH/bin`, so be sure this it is correctly
added to you path (like with `export PATH=$PATH:/Users/david/go/bin`).

Create a private key (in the d-voting root folder):

```
crypto bls signer new --save private.key
```

Copy the private key from the d-voting root folder to the cli/memcoin folder:

```
cp private.key cli/memcoin/
```

Install memcoin (this requires the private key in cli/memcoin):

```
cd cli/memcoin
go install
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
rm -rf /tmp/node{1,2,3}
```

# Testing
## Automate the memcoin setup using tmux
If you have `tmux` installed, you can start a `tmux` session that will execute the above setup by running:
```
./test.sh
```
You can move around the panes by pressing `Ctrl+B` followed by an arrow key.

To end the session, in a different terminal run:
```
tmux kill-session -t d-voting-test
```
or kill each pane then the session with `Ctrl+D`.

You'll have to delete the state files manually (see "Run the nodes").

## Run the scenario test

If nodes are running and `setup.sh` has been called, you can run a test
scenario:

```
LLVL=info memcoin --config /tmp/node1 e-voting scenarioTest
```

# Use the frontend

See README in `web/`.
