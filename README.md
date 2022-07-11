<div align="center">

<img width="300px" src="docs/assets/logo-white.png#gh-dark-mode-only"/>
<img width="300px" src="docs/assets/logo-black.png#gh-light-mode-only"/>

<p></p>

<table>
<tr>
    <td>Global</td>
    <td>
        <a href="https://sonarcloud.io/summary/new_code?id=dedis_d-voting">
            <img src="https://sonarcloud.io/api/project_badges/measure?project=dedis_d-voting&metric=alert_status">
        </a>
        <a href="https://github.com/dedis/d-voting/graphs/contributors">
            <img alt="GitHub contributors" src="https://img.shields.io/github/contributors/dedis/d-voting">
        </a>
        <a href="https://github.com/dedis/d-voting/releases">
            <img alt="GitHub release (latest SemVer)" src="https://img.shields.io/github/v/release/dedis/d-voting">
        </a>
      </td>
</tr>
<tr>
    <td>Blockchain</td>
    <td>
        <a href="https://github.com/dedis/d-voting/actions/workflows/go_test.yml">
            <img src="https://github.com/dedis/d-voting/actions/workflows/go_test.yml/badge.svg">
        </a>
        <a href="https://github.com/dedis/d-voting/actions/workflows/go_memcoin_test.yml">
            <img src="https://github.com/dedis/d-voting/actions/workflows/go_memcoin_test.yml/badge.svg">
        </a>
        <a href="https://github.com/dedis/d-voting/actions/workflows/go_integration_tests.yml">
            <img src="https://github.com/dedis/d-voting/actions/workflows/go_integration_tests.yml/badge.svg">
        </a><br/>
        <a href="https://coveralls.io/github/dedis/d-voting?branch=main">
            <img src="https://coveralls.io/repos/github/dedis/d-voting/badge.svg?branch=main">
        </a>
        <a href="https://goreportcard.com/report/github.com/dedis/d-voting">
            <img src="https://goreportcard.com/badge/github.com/dedis/d-voting">
        </a>
        <a href="https://pkg.go.dev/github.com/dedis/d-voting">
            <img src="https://pkg.go.dev/badge/github.com/dedis/d-voting.svg" alt="Go Reference">
        </a>
    </td>
<tr>
<tr>
    <td>WEB</td>
    <td>
        <a href="https://github.com/dedis/d-voting/actions/workflows/web_frontend_lint.yml">
            <img src="https://github.com/dedis/d-voting/actions/workflows/web_frontend_lint.yml/badge.svg">
        </a>
        <a href="https://github.com/dedis/d-voting/actions/workflows/web_backend_lint.yml">
            <img src="https://github.com/dedis/d-voting/actions/workflows/web_backend_lint.yml/badge.svg">
        </a>
    </td>
</tr>
</table>

</div>

# D-Voting

**D-Voting** is an e-voting platform based on the
[Dela](https://github.com/dedis/dela) blockchain. In short:

- An open platform to run voting instances on a blockchain
- Provides privacy of votes with state-of-the art protocols
- Fully auditable and decentralized process

## Global architecture

Find more about the architecture on the [documentation
website](https://dedis.github.io/d-voting/#/).

![Global component diagram](http://www.plantuml.com/plantuml/proxy?src=https://raw.githubusercontent.com/dedis/d-voting/main/docs/assets/component-global.puml)


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

Additionally, you can build the memcoin binary with:

```sh
go build ./cli/memcoin
```

# Run the nodes

In three different terminal sessions, from the root folder:

```sh
pk=adbacd10fdb9822c71025d6d00092b8a4abb5ebcb673d28d863f7c7c5adaddf3

LLVL=info memcoin --config /tmp/node1 start --postinstall \
  --promaddr :9100 --proxyaddr :9080 --proxykey $pk --listen tcp://0.0.0.0:2001 --public //localhost:2001

LLVL=info memcoin --config /tmp/node2 start --postinstall \
  --promaddr :9101 --proxyaddr :9081 --proxykey $pk --listen tcp://0.0.0.0:2002 --public //localhost:2002

LLVL=info memcoin --config /tmp/node3 start --postinstall \
  --promaddr :9102 --proxyaddr :9082 --proxykey $pk --listen tcp://0.0.0.0:2003 --public //localhost:2003
```

Then you should be able to run the setup script:

```sh
./setup.sh
```
With this other script you can choose the number of nodes that you want to set up:

```sh
./setupnNode.sh -n 3 
```

This script will setup the nodes and services. If you restart do not forget to
remove the old state:

```sh
rm -rf /tmp/node{1,2,3}
```

# Testing

## Automate the previous setup using `tmux`

If you have `tmux` installed, you can start a `tmux` session that will
execute the above setup by running in the project root `./runNode.sh -n 3`. This
command takes as argument the number of nodes. 
Once the session is started, you can move around the panes with
`Ctrl+B` followed by arrow keys or by `N`. You can also have an overview of the windows 
with `Ctrl+B` followed by `S`.


To end the session, run `./kill_test.sh`,
which will kill each window then the `tmux` session (which you can do manually with `Ctrl+D`),
then delete the node data (i.e. the files `/tmp/node{1,2,3}`).

## Run the scenario test

If nodes are running and `setup.sh` or `./setupnNode.sh -n 3` has been called, you can run a test
scenario:

```sh
sk=28912721dfd507e198b31602fb67824856eb5a674c021d49fdccbe52f0234409
LLVL=info memcoin --config /tmp/node1 e-voting scenarioTest --secretkey $sk
```

You can also run scenario_test.go, by running in the integration folder this command:
```sh
NNODES=3 go test -v scenario_test.go
```


For reference, here is a hex-encoded kyber Ed25519 keypair:

Public key: `adbacd10fdb9822c71025d6d00092b8a4abb5ebcb673d28d863f7c7c5adaddf3`

Secret key: `28912721dfd507e198b31602fb67824856eb5a674c021d49fdccbe52f0234409`

## Run the scenario test with docker 
Use the following commands to launch and set up nodes, and start the scenario test with user defined number of nodes.

First build the docker image `docker build -t node .`

Afterwards use the following commands, replace 4 by the desired nb of nodes :

```sh
./runNode.sh -n 4 -a true -d true
./setupnNode.sh -n 4 -d true

NNODES=4 KILLNODE=true go test -v -run ^TestScenario$ github.com/dedis/d-voting/integration -count=1
```

Here we set KILLNODE=true or false to decide whether kill and restart a node during the election process. By default, it's set to false.

To end the session, run `./kill_test.sh`.

To launch multiple test and get statistics, run `./autotest.sh -n 10 -r 15`.

N.B. run following commands to get help
```sh
./runNode.sh -h
./setupnNode.sh -h
./autotest.sh -h
```


# Use the frontend

See README in `web/`.

# Debian deployment

A package registry with debian packages is available at http://apt.dedis.ch.
To install a package run the following:

```sh
echo "deb http://apt.dedis.ch/ squeeze main" >> /etc/apt/sources.list
wget -q -O- http://apt.dedis.ch/unicore.gpg | sudo apt-key add -
sudo apt update
sudo apt install dedis-dvoting
```

# Metrics

A d-Voting node exposes Prometheus metrics. You can start an HTTP server that
serves those metrics with:

```sh
./memcoin --config /tmp/node1 metrics start --addr 127.0.0.1:9100 --path /metrics
```

Build info can be added to the binary with the `ldflags`, at build time. Infos
are stored on variables in the root `mod.go`. For example:

```sh
versionFlag="github.com/dedis/d-voting.Version=`git describe --tags`"
timeFlag="github.com/dedis/d-voting.BuildTime=`date +'%d/%m/%y_%H:%M'`"

go build -ldflags="-X $versionFlag -X $timeFlag" ./cli/memcoin
```

Note that `make build` will do that for you.
