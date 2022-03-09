<div align="center">

<img width="300px" src="docs/assets/logo-white-bg.png"/>

<table class="tg">
<tbody>
  <tr>
    <th class="tg-amwm" colspan="2">Global</th>
  </tr>
  <tr>
    <td class="tg-baqh" colspan="2">
        <a href="https://sonarcloud.io/summary/new_code?id=dedis_d-voting">
            <img src="https://sonarcloud.io/api/project_badges/measure?project=dedis_d-voting&metric=alert_status">
        </a>
    </td>
  </tr>
  <tr>
    <th class="tg-amwm" colspan="2">Blockchain</th>
  </tr>
  <tr>
    <td class="tg-baqh">Tests</td>
    <td class="tg-baqh">Quality</td>
  </tr>
  <tr>
    <td class="tg-baqh">
        <a href="https://github.com/dedis/d-voting/actions/workflows/go_test.yml">
            <img src="https://github.com/dedis/d-voting/actions/workflows/go_test.yml/badge.svg">
        </a>
        <a href="https://github.com/dedis/d-voting/actions/workflows/go_integration_tests.yml">
            <img src="https://github.com/dedis/d-voting/actions/workflows/go_integration_tests.yml/badge.svg">
        </a>
        <a href="https://github.com/dedis/d-voting/actions/workflows/go_memcoin_test.yml">
            <img src="https://github.com/dedis/d-voting/actions/workflows/go_memcoin_test.yml/badge.svg">
        </a>
    </td>
    <td class="tg-baqh">
        <a href="https://coveralls.io/github/dedis/d-voting?branch=main">
            <img src="https://coveralls.io/repos/github/dedis/d-voting/badge.svg?branch=main">
        </a>
    </td>
  </tr>
  <tr>
    <th class="tg-amwm" colspan="2">Web client</th>
  </tr>
  <tr>
    <td class="tg-baqh">Frontend</td>
    <td class="tg-baqh">Backend</td>
  </tr>
  <tr>
    <td class="tg-baqh">
        <a href="https://github.com/dedis/d-voting/actions/workflows/web_frontend_lint.yml">
            <img src="https://github.com/dedis/d-voting/actions/workflows/web_frontend_lint.yml/badge.svg">
        </a>
    </td>
    <td class="tg-baqh">
        <a href="https://github.com/dedis/d-voting/actions/workflows/web_backend_lint.yml">
            <img src="https://github.com/dedis/d-voting/actions/workflows/web_backend_lint.yml/badge.svg">
        </a>
    </td>
  </tr>
</tbody>
</table>

</div>

# D-Voting

**D-Voting** is an e-voting platform based on the
[Dela](https://github.com/dedis/dela) blockchain. In short:

- An open platform to run voting instances on a blockchain
- Provides privacy of votes with state-of-the art protocols
- Fully auditable and decentralized process

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

# Metrics

A d-Voting node exposes Prometheus metrics. You can start an HTTP server that
serves those metrics with:

```sh
./memcoin --config /tmp/node1 metrics start --addr 127.0.0.1:9100 --path /metrics
```

Build info can be added to the binary with the `ldflags`, at build time.. Infos
are stored on variables in the root `mod.go`. For example:

```sh
versionFlag="github.com/dedis/d-voting.Version=`git describe --tags`"
timeFlag="github.com/dedis/d-voting.BuildTime=`date +'%d/%m/%y_%H:%M'`"

go build -ldflags="-X $versionFlag -X $timeFlag" ./cli/memcoin
```

