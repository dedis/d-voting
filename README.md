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
        <a href="https://github.com/dedis/d-voting/actions/workflows/go_release.yml">
            <img src="https://github.com/dedis/d-voting/actions/workflows/go_release.yml/badge.svg">
        </a><br/>
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
        <a href="https://github.com/dedis/d-voting/actions/workflows/go_dvoting_test.yml">
            <img src="https://github.com/dedis/d-voting/actions/workflows/go_dvoting_test.yml/badge.svg">
        </a>
        <a href="https://github.com/dedis/d-voting/actions/workflows/go_scenario_test.yml">
            <img src="https://github.com/dedis/d-voting/actions/workflows/go_scenario_test.yml/badge.svg">
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
[Dela](https://github.com/dedis/dela) blockchain. It uses state-of-the-art
protocols that guarantee privacy of votes and a fully decentralized process.
This project was born in early 2021 and has been iteratively implemented by EPFL
students under the supervision of DEDIS members.

⚠️ This project is still under development and should not be used for real
forms.

Main properties of the system are the following:

<div align="center">
    <img height="45px" src="docs/assets/spof-black.png#gh-light-mode-only">
    <img height="45px" src="docs/assets/spof-white.png#gh-dark-mode-only">
</div>

**No single point of failure** - The system is supported by a decentralized
network of blockchain nodes, making no single party able to break the system
without compromising a Byzantine threshold of nodes. Additionally,
side-protocols always distribute trust among nodes: The distributed key
generation protocol (DKG) ensures that a threshold of honest node is needed to
decrypt ballots, and the shuffling protocol needs at least one honest node to
ensure privacy of voters. Only the identification and authorization mechanism
make use of a central authority, but can accommodate to other solutions.

<div align="center">
    <img height="45px" src="docs/assets/privacy-black.png#gh-light-mode-only">
    <img height="45px" src="docs/assets/privacy-white.png#gh-dark-mode-only">
</div>

**Privacy** - Ballots are cast on the client side using a safely-held
distributed key-pair. The private key cannot not be revealed without coercing a
threshold of nodes, and voters can retrieve the public key on any node. Ballots
are decrypted only once a cryptographic process ensured that cast ballots cannot
be linked to the original voter.

<div align="center">
    <img height="50px" src="docs/assets/audit-black.png#gh-light-mode-only">
    <img height="50px" src="docs/assets/audit-white.png#gh-dark-mode-only">
</div>

**Transparency/Verifiability/Auditability** - The whole voting process is
recorded on the blockchain and signed by a threshold of blockchain nodes. Anyone
can read and verify the log of events stored on the blockchain. Malicious
behavior can be detected, voters can check that ballots are cast as intended,
and auditors can witness the voting process.

## 🧩 Global architecture

The project has 4 main high-level components:

**Proxy** - A proxy offers the mean for an external actor such as a website to
interact with a blockchain node. It is a component of the blockchain node that
exposes HTTP endpoints for external entities to send commands to the node. The
proxy is notably used by the web clients to use the voting system.

**Web frontend** - The web frontend is a web app built with React. It offers a
view for end-users to use the D-Voting system. The app is meant to be used by
voters and admins. Admins can perform administrative tasks such as creating an
form, closing it, or revealing the results. Depending on the task, the web
frontend will directly send HTTP requests to the proxy of a blockchain node, or
to the web-backend.

**Web backend** - The web backend handles authentication and authorization. Some
requests that need specific authorization are relayed from the web-frontend to
the web-backend. The web backend checks the requests and signs messages before
relaying them to the blockchain node, which trusts the web-backend. The
web-backend has a local database to store configuration data such as
authorizations. Admins use the web-frontend to perform updates.

**Blockchain node** - A blockchain node is the wide definition of the program
that runs on a host and participate in the voting logic. The blockchain node
is built on top of Dela with an additional d-voting smart contract, proxy, and
two services: DKG and verifiable Shuffling. The blockchain node is more
accurately a subsystem, as it wraps many other components. Blockchain nodes
communicate through gRPC with the [minogrpc][minogrpc] network overlay. We
sometimes refer to the blockchain node simply as a "node".

The following component diagrams summarizes the interaction between those
high-level components:

[minogrpc]: https://github.com/dedis/dela/tree/master/mino/minogrpc

![Global component diagram](http://www.plantuml.com/plantuml/proxy?src=https://raw.githubusercontent.com/dedis/d-voting/main/docs/assets/component-global.puml)

You can find more information about the architecture on the [documentation
website](https://dedis.github.io/d-voting/#/).

## Workflow

A form follows a specific workflow to ensure privacy of votes. Once an
form is created and open, there are 4 main steps from the cast of a ballot
to getting the result of the form:

<div align="center">
    <img height="55px" src="docs/assets/encrypt-black.png#gh-light-mode-only">
    <img height="55px" src="docs/assets/encrypt-white.png#gh-dark-mode-only">
</div>

**1) Create ballot** The voter gets the shared public key and encrypts locally
its ballot. The shared public key can be retrieved on any node and is associated
to a private key that is distributed among the nodes. This process is done on
the client's browser using the web-frontend.

<div align="center">
    <img height="80px" src="docs/assets/cast-black.png#gh-light-mode-only">
    <img height="80px" src="docs/assets/cast-white.png#gh-dark-mode-only">
</div>

**2) Cast ballot** The voter submits its encrypted ballot as a transaction to one
of the blockchain node. This operation is relayed by the web-backend which
verifies that the voters has the right to vote. If successful, the encrypted
ballot is stored on the blockchain. At this stage each encrypted ballot is
associated to its voter on the blockchain.

<div align="center">
    <img height="70px" src="docs/assets/shuffle-black.png#gh-light-mode-only">
    <img height="70px" src="docs/assets/shuffle-white.png#gh-dark-mode-only">
</div>

**3) Shuffle ballots** Once the form is closed by an admin, ballots are
shuffled to ensure privacy of voters. This operation is done by a threshold of
node that each perform their own shuffling. Each shuffling guarantees the
integrity of ballots while re-encrypting and changing the order of ballots. At
this stage encrypted ballots cannot be linked back to their voters.

<div align="center">
    <img height="90px" src="docs/assets/reveal-black.png#gh-light-mode-only">
    <img height="90px" src="docs/assets/reveal-white.png#gh-dark-mode-only">
</div>

**4) Reveal ballots** Once ballots have been shuffled, they are decrypted and
revealed. This operation is done only if the previous step is correctly
executed. The decryption is done by a threshold of nodes that must each provide
a contribution to achieve the decryption. Once done, the result of the form
is stored on the blockchain.

For a more formal and in-depth overview of the workflow, see the
[documentation](https://dedis.github.io/d-voting/#/api?id=signed-requests)

## Smart contract

A smart contract is a piece of code that runs on a blockchain. It defines a set
of operations that act on a global state (think of it as a database) and can be
triggered with transactions. What makes a smart contract special is that its
executions depends on a consensus among blockchain nodes where operations are
successful only if a consensus is reached. Additionally, transactions and their
results are permanently recorded and signed on an append-only ledger, making any
operations on the blockchain transparent and permanent.

In the D-Voting system a single D-Voting smart contract handles the forms.
The smart contract ensures that forms follow a correct workflow to
guarantees its desirable properties such as privacy. For example, the smart
contract won't allow ballots to be decrypted if they haven't been previously
shuffled by a threshold of nodes.

## Services

Apart from executing smart contracts, blockchain nodes need additional side
services to support a form. Side services can read from the global state
and send transactions to write to it via the D-Voting smart contract. They are
used to perform specific protocol executions not directly related to blockchain
protocols such as the distributed key generation (DKG) and verifiable shuffling
protocols.

### Distributed Key Generation (DKG)

The DKG service allows the creation of a distributed key-pair among multiple
participants. Data encrypted with the key-pair can only be decrypted with the
contribution of a threshold of participants. This makes it convenient to
distribute trust on encrypted data. In the D-Voting project we use the Pedersen
[[1]] version of DKG.

The DKG service needs to be setup at the beginning of each new form, because
we want each form to have its own key-pair. Doing the setup requires two
steps: 1\) Initialization and 2\) Setup. The initialization creates new RPC
endpoints on each node, which they can use to communicate with each other. The
second step, the setup, must be executed on one of the node. The setup step
starts the DKG protocol and generates the key-pair. Once done, the D-Voting
smart contract can be called to open the form, which will retrieve the DKG
public key and save it on the smart contract.

[1]: https://dl.acm.org/doi/10.5555/1754868.1754929

### Verifiable shuffling

The shuffling service ensures that encrypted votes can not be linked to the user
who cast them. Once the service is setup, each node can perform what we call a
"shuffling step". A shuffling step re-orders an array of elements such that the
integrity of the elements is guarantee (i.e no elements have been modified,
added, or removed), but one can't trace how elements have been re-ordered.

In D-Voting we use the Neff [[2]] implementation of verifiable shuffling. Once
a form is closed, an admin can trigger the shuffling steps from the nodes.
During this phase, every node performs a shuffling on the current list of
encrypted ballots and tries to submit it to the D-Voting smart contract. The
smart contract will accept only one shuffling step per block in the blockchain.
Nodes re-try to shuffle the ballots, using the latest shuffled list in the
blockchain, until the result of their shuffling has been committed to the
blockchain or a threshold of nodes successfully submitted their own shuffling
results.

[2]: https://dl.acm.org/doi/10.1145/501983.502000

## 📁 Folders structure

<pre>
<code>
.
├── cli    
│   ├── cosipbftcontroller  Custom initialization of the blockchain node
│   ├── <b>dvoting</b>             Build the node CLI
│   └── postinstall         Custom node CLI setup
├── <b>contracts</b>           
│   └── <b>evoting</b>             D-Voting smart contract
│       └── controller      CLI commands for the smart contract
├── deb-package             Debian package for deployment
├── docs                    Documentation 
├── integration             Integration tests
├── internal                Internal packages: testing, tooling, tracing
├── metrics             
│   └── controller          CLI commands for Prometheus
├── proxy                   Defines and implements HTTP handlers for the REST API
├── <b>services</b>
│   ├── dkg  
│   │   └── <b>pedersen</b>        Implementation of the DKG service
│   └── shuffle   
│       └── <b>neff</b>            Implementation of the shuffle service
└── <b>web</b>
    ├── <b>backend</b>
    │   └── src             Sources of the web backend (express.js server)
    └── <b>frontend</b>
        └── src             Sources of the web frontend (react app)
</code>
</pre>

## 👩‍💻👨‍💻 Contributors

<table>
    <tr style="font-weight:bold">
        <td>Period</td>
        <td>Contributors(s)</td>
        <td>Activities</td>
        <td>Links</td>
    </tr>
    <tr>
        <td>Spring 2021</td>
        <td>Students: Anas Ibrahim, Vincent Parodi<br>Supervisor: Noémien Kocher</td>
        <td>Initial implementation of the smart contract and services</td>
        <td><a href="https://www.epfl.ch/labs/dedis/wp-content/uploads/2021/07/report-2021-1-Vincent-Anas_EvotingDela.pdf">Report</a></td>
    </tr>
    <tr>
        <td>Spring 2021</td>
        <td>Student: Sarah Antille<br>Supervisor: Noémien Kocher</td>
        <td>Initial implementation of the web frontend in react</td>
        <td><a href="https://www.epfl.ch/labs/dedis/wp-content/uploads/2021/07/report-2021-1-SarahAntille_EvotingonDela.pdf">Report</a></td>
    </tr>
    <tr>
        <td>Fall 2021</td>
        <td>Students: Auguste Baum, Emilien Duc<br>Supervisor: Noémien Kocher</td>
        <td>Adds a flexible form structure. Improves robustness and security.</td>
        <td>
            <a href="https://www.epfl.ch/labs/dedis/wp-content/uploads/2022/02/report-2021-3-baum-auguste-Dvoting.pdf">Report</a>,
            <a href="https://www.epfl.ch/labs/dedis/wp-content/uploads/2022/02/presentation-2021-3-baum-auguste-dvoting.pdf">Presentation</a>
        </td>
    </tr>
    <tr>
        <td>Fall 2021</td>
        <td>Students: Ambroise Borbely<br>Supervisor: Noémien Kocher</td>
        <td>Adds authentication and authorization mechanism on the frontend.</td>
        <td><a href="https://www.epfl.ch/labs/dedis/wp-content/uploads/2022/02/report-2021-3-borbely-ambroise-DVotingFrontend.pdf">Report</a></td>
    </tr>
    <tr>
        <td>Spring 2022</td>
        <td>Students: Guanyu Zhang, Igowa Giovanni<br>Supervisor: Noémien Kocher<br>Assistant: Emilien Duc</td>
        <td>Improves production-readiness: deploy a test pipeline and analyze the system's robustness.</td>
        <td>
            <a href="https://www.epfl.ch/labs/dedis/wp-content/uploads/2022/07/report-2022-1-giovanni-zhang-d-voting-production.pdf">Report</a>,
            <a href="https://www.epfl.ch/labs/dedis/wp-content/uploads/2022/07/presentation-2022-1-giovanni-zhang-d-voting-production.pdf">Presentation</a>
        </td>
    </tr>
    <tr>
        <td>Spring 2022</td>
        <td>Students: Badr Larhdir, Capucine Berger<br>Supervisor: Noémien Kocher</td>
        <td>Major iteration over the frontend - design and functionalities: implements a flexible form form, nodes setup, and result page.</td>
        <td>
            <a href="https://www.epfl.ch/labs/dedis/wp-content/uploads/2022/07/report-2022-1-capucine-badr-d-voting-frontend.pdf">Report</a>,
            <a href="https://www.epfl.ch/labs/dedis/wp-content/uploads/2022/07/presentation-2022-1-capucine-badr-d-voting-frontend.pdf">Presentation</a>
        </td>
    </tr> 
    <tr>
        <td>Fall 2022</td>
        <td>Students: Amine Benaziz, Albert Troussard<br>Supervisors: Noémien Kocher, Pierluca Borso<br>Assistant: Emilien Duc</td></td>
        <td>Improves production-readiness: implement a new transaction system, improves the setup script, analyze performance and robustness with improved testing.</td>
        <td>
            <a href="https://www.epfl.ch/labs/dedis/wp-content/uploads/2023/01/report-2022-3-Amine-Benaziz-Albert-Troussard-D_Voting_Production_Readiness.pdf">Report</a>, <a href="https://www.epfl.ch/labs/dedis/wp-content/uploads/2023/01/presentation-2022-3-chen-ahmed-ghita-khadija-amine-albert_d_voting.pdf">Presentation (common)
        </td>
    </tr>
    <tr>
        <td>Fall 2022</td>
        <td>Students: Ahmed Elalamy, Ghita Tagemouati, Khadija Tagemouati<br>Supervisor: Noémien Kocher</td></td>
        <td>Implements a policy-based authentication system, i18n of forms, individual results, and other minor improvement on the usability.</td>
        <td>
            <a href="https://www.epfl.ch/labs/dedis/wp-content/uploads/2023/01/report_2022-3-ghita_khadija-ahmed-d_voting_frontend.pdf">Report</a>, <a href="https://www.epfl.ch/labs/dedis/wp-content/uploads/2023/01/presentation-2022-3-chen-ahmed-ghita-khadija-amine-albert_d_voting.pdf">Presentation (common)
        </td>
    </tr>
    <tr>
        <td>Fall 2022</td>
        <td>Students: Chen Chang Lew<br>Supervisors: Noémien Kocher, Pierluca Borso, Simone Colombo</td></td>
        <td>Threat modelling of the system, security audit: identification of threats and technical debts, design of a vote verifiability feature.</td>
        <td>
            <a href="https://www.epfl.ch/labs/dedis/wp-content/uploads/2023/01/report-2022-3-chenChangLew-Dvotingsecurityreport2.pdf">Report</a>,
            <a href="https://www.epfl.ch/labs/dedis/wp-content/uploads/2023/01/report-2022-3-chenChangLew-DvotingSecurityReport1.pdf">Security audit</a>, <a href="https://www.epfl.ch/labs/dedis/wp-content/uploads/2023/01/presentation-2022-3-chen-ahmed-ghita-khadija-amine-albert_d_voting.pdf">Presentation (common)
        </td>
    </tr>
</table>

# ⚙️ Setup

1: Install [Go](https://go.dev/dl/) (at least 1.19).

2: Install the `crypto` utility from Dela:

```sh
git clone https://github.com/dedis/dela.git
cd dela/cli/crypto
go install
```

Go will install the binaries in `$GOPATH/bin`, so be sure this it is correctly
added to you path (like with `export PATH=$PATH:/Users/david/go/bin`).

3: [Install tmux](https://github.com/tmux/tmux/wiki/Installing)

4: The authorization are stored in a postgres database. We use Docker-compose to define it easily and allow us to start and stop all the services with a single command. You will then need to install docker compose using the command :

```sh
sudo snap install docker
```

And finish the installation by following the [steps](https://www.digitalocean.com/community/tutorial_collections/how-to-install-docker-compose) depending on your OS.

# Setup a simple system with 5 nodes (Linux and MacOS)

If you are using Windows and cannot use tmux, you need to follow the instructions in [this](<#Setup-a-simple-system-with-5-nodes-(Windows)>)
section.

1: Only for the first time

```sh
cd web/backend
npm install
cp config.env.template config.env
cd ../frontend
npm install
cd ../..
```

2: In a new window, you will need to start the database:

```sh
cd d-voting/web/backend/src
sudo docker-compose up
```

If you want to stop the services you can use the following command:

```sh
sudo docker-compose down
```

If you want to have a way to check the database you can install [pgAdmin](https://www.pgadmin.org/download/).

3: Then run the following script to start and setup the nodes and the web server:

```sh
cd d-voting
./runSystems.sh -n 5
```

This will run 8 terminal sessions. You can navigate by hitting
<kbd>CTRL</kbd>+<kbd>B</kbd> and then <kbd>S</kbd>. Use the arrows to select a
window.

4: Stop nodes
If you want to stop the system, you can use the following command:

(If you forgot, this will be done automatically when you start a new system)

```sh
./kill_test.sh
```

5: Troubleshoot

If while running

```sh
./runSystems.sh -n 5
```

You get this error:

```sh
Error: listen EADDRINUSE: address already in use :::5000
```

then in the file runSystems.sh, replace the line:

```sh
tmux send-keys -t $s:{end} "cd web/backend && npm start" C-m
```

with

```sh
tmux send-keys -t $s:{end} "cd web/backend && PORT=4000 npm start" C-m
#or any other available port
```

And in the web/frontend/src/setupProxy.js file, change :

```sh
target: 'http://localhost:5000',
```

with

```sh
target: 'http://localhost:4000',
```

# Setup a simple system with 3 nodes (Windows)

In three different terminal sessions, from the root folder:

```sh
pk=adbacd10fdb9822c71025d6d00092b8a4abb5ebcb673d28d863f7c7c5adaddf3

LLVL=info dvoting --config /tmp/node1 start --postinstall \
  --promaddr :9100 --proxyaddr :9080 --proxykey $pk --listen tcp://0.0.0.0:2001 --public //localhost:2001

LLVL=info dvoting --config /tmp/node2 start --postinstall \
  --promaddr :9101 --proxyaddr :9081 --proxykey $pk --listen tcp://0.0.0.0:2002 --public //localhost:2002

LLVL=info dvoting --config /tmp/node3 start --postinstall \
  --promaddr :9102 --proxyaddr :9082 --proxykey $pk --listen tcp://0.0.0.0:2003 --public //localhost:2003
```

Then you should be able to run the setup script:

```sh
./setup.sh
```

With this other script using tmux you can choose the number of nodes that you
want to set up:

```sh
./setupnNode.sh -n 3
```

This script will setup the nodes and services. If you restart do not forget to
remove the old state:

```sh
rm -rf /tmp/node{1,2,3}
```

3: Launch the web backend

From a new terminal session, run:

```sh
cd web/backend
# if this is the first time, run `npm install` and `cp config.env.template config.env` first
npm start
```

4: Launch the web frontend

From a new terminal session, run:

```sh
cd web/frontend
# if this is the first time, run `npm install` first
REACT_APP_PROXY=http://localhost:9081 REACT_APP_NOMOCK=on npm start
```

Note that you need to be on EPFL's network to login with Tequila. Additionally,
once logged with Tequila, update the redirect URL and replace
`dvoting-dev.dedis.ch` with `localhost`. Once logged, you can create an
form.

# Testing

## Run the scenario test

If nodes are running and `setup.sh` or `./runSystem.sh -n 3 --backend false --frontend false` (for this test you don't want the user interface so the web components are not needed) has been called,
you can run a test scenario:

```sh
sk=28912721dfd507e198b31602fb67824856eb5a674c021d49fdccbe52f0234409
LLVL=info dvoting --config /tmp/node1 e-voting scenarioTest --secretkey $sk
```

You can also run scenario_test.go, by running in the integration folder this
command:

```sh
NNODES=3 go test -v scenario_test.go
```

For reference, here is a hex-encoded kyber Ed25519 keypair:

Public key: `adbacd10fdb9822c71025d6d00092b8a4abb5ebcb673d28d863f7c7c5adaddf3`

Secret key: `28912721dfd507e198b31602fb67824856eb5a674c021d49fdccbe52f0234409`

<!---
Currently not working

## Run the scenario test with docker

Use the following commands to launch and set up nodes, and start the scenario
test with user defined number of nodes.

First build the docker image `docker build -t node .`

Afterwards use the following commands, replace 4 by the desired nb of nodes :

```sh
./runNode.sh -n 4 -a true -d true
./setupnNode.sh -n 4 -d true

NNODES=4 KILLNODE=true go test -v -run ^TestScenario$ github.com/dedis/d-voting/integration -count=1
```

Here we set KILLNODE=true or false to decide whether kill and restart a node
during the voting process. By default, it's set to false.

To end the session, run `./kill_test.sh`.

To launch multiple test and get statistics, run `./autotest.sh -n 10 -r 15`.

N.B. run following commands to get help

```sh
./runNode.sh -h
./setupnNode.sh -h
./autotest.sh -h
```

-->

# Use the frontend

See README in `web/`.

# Debian deployment

A package registry with debian packages is available at http://apt.dedis.ch.
To install a package run the following:

```sh
echo "deb http://apt.dedis.ch/ squeeze main" >> /etc/apt/sources.list
wget -q -O- http://apt.dedis.ch/dvoting-release.pgp | sudo apt-key add -
sudo apt update
sudo apt install dedis-dvoting
```

# Metrics

A d-Voting node exposes Prometheus metrics. You can start an HTTP server that
serves those metrics with:

```sh
./dvoting --config /tmp/node1 metrics start --addr 127.0.0.1:9100 --path /metrics
```

Build info can be added to the binary with the `ldflags`, at build time. Infos
are stored on variables in the root `mod.go`. For example:

```sh
versionFlag="go.dedis.ch/d-voting.Version=`git describe --tags`"
timeFlag="go.dedis.ch/d-voting.BuildTime=`date +'%d/%m/%y_%H:%M'`"

go build -ldflags="-X $versionFlag -X $timeFlag" ./cli/dvoting
```

Note that `make build` will do that for you.

# Benchmarks

For more details, see https://github.com/c4dt/d-voting/issues/47

```
+------------+-------------+--------------+
I            I   1'000     I   10'000     I
+------------+-------------+--------------+
I CASTING    I  95s        I  888s        I
I            I  95ms/vote  I   95ms/vote  I
+------------+-------------+--------------+
I   Blocks   I 179         I 1466         I
I            I   5.6/block I    6.8/block I
+------------+-------------+--------------+
I SHUFFLING  I  56s        I  542s        I
+------------+-------------+--------------+
I   Blocks   I   9         I   10         I
+------------+-------------+--------------+
I DECRYPTING I   5s        I   49s        I
+------------+-------------+--------------+
I   Blocks   I   2         I    3         I
+------------+-------------+--------------+
I COMBINING  I   1s        I   47s        I
+------------+-------------+--------------+
I   Blocks   I   1         I    1         I
+------------+-------------+--------------+
I DISPLAYING I  <1s        I    5s        I
+------------+-------------+--------------+

```

- 10'000 votes
  - casting: 95s, 179 blocks - 95ms/vote, 5.6 votes / block

---

<img width="200px" src="docs/unicore_logo.png"/>

This project has received funding from the European Union's Horizon 2020 
research and innovation programme under grant agreement No 825377.
