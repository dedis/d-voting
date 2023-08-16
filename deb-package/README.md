# Run the web backend

You need to have node 18.x version.

1) Install dependencies and transpile:

```sh
cd web/frontend
npm install
NODE_ENV=production ./node_modules/.bin/tsc --outDir build
```

2) Copy `node_modules`, `config.env`, and `dbUtil.js` somewhere.

3) Run the web backend:

```sh
source x/y/config.env && \
    NODE_ENV=production PORT=6000 NODE_PATH=x/y/node_modules \
    /usr/bin/node /x/y/build/Server.js
```

`Server.js` is the result of the transpilation. Replace locations appropriately
and be sure to *export* variables in `config.env` (i.e do `export xx=yy`).
However it is recommended to use a service manager such as systemd to run the
app.

4) Use the CLI to set yourself up as an admin 

```sh
npx cli addAdmin --sciper 1234
```

# Run the web frontend

You need to have node 18.x version.

1) Install dependencies and transpile:

```sh
cd web/frontend
npm install
NODE_ENV=production HTTPS=true BUILD_PATH=x/y/build ./node_modules/.bin/react-scripts build
```

2) Configure an HTTP server

You should server the `x/y/build` folder and proxy all requests on `/api` to the
web backend. For example with nginx:

```
    location / {
        root   /x/y/build;
        index  index.html;
        autoindex on;
        try_files $uri /index.html;
    }

    location /api {
        proxy_pass http://127.0.0.1:6000;
    }
```

# Configure a network of nodes

## Run a node

Check how a node is started in `pkg/opt/dedis/dvoting/bin/start-voting`. For
example:

```sh
LLVL=info ./memcoin-darwin-amd64-v0_4_5-alpha \
    --config /tmp/node8 \
    start \
    --postinstall \
    --promaddr 0.0.0.0:9101 \
    --proxyaddr 0.0.0.0:9080 \
    --listen //0.0.0.0:9000 \
    --public //localhost:9000 \
    --routing flat \
    --noTLS=true \
    --proxykey adbacd10fdb9822c71025d6d00092b8a4abb5ebcb673d28d863f7c7c5adaddf3
```

## Network config

Ensure that the public address is correct. For instance, in
`/etc/dedis/dvoting/config.env`, replace:

```sh
export dela_public="//localhost:9000"
```

with the node's public address:

```sh
export dela_public="//172.16.253.150:9000"
```

and don't forget to restart the service!

```sh
service d-voting restart
```

## Create a roster

If you keep TLS encryption at the gRPC level, you must share the certificates
between the nodes. To do that you generate credentials on the first node, and
make all other nodes send their certificates to the first node using the
credentials.

**generate credentials** (on the first node):

Get the token and certificate (24h * 30 = 720):

```sh
sudo memcoin --config /var/opt/dedis/dvoting/data/dela minogrpc token \
    --expiration 720h
```

This result, which looks like as follow, will be given to node's operators:

```
--token b6VhdQEPXKOtZHpng8E8jw== --cert-hash oNeyrA864P2cP+TT6IE6GvkeEI/Ec4rOlZWEWiQkQKk=
```

**share certificates** (all other nodes):

Join the network. This operation will make the node share its certificate to the
MASTER node, which, in turn, will share its known certificates to the node. Note
that the certificates are stored in the DB, which means that this operation must
be re-done in case the DB is reset.

```sh
sudo memcoin --config /var/opt/dedis/dvoting/data/dela minogrpc join \
    --address <MASTER NODE ADDRESS> --token <TOKEN> --cert-hash <CERT HASH>
```

Example of `<MASTER NODE ADDRESS>`: `'//172.16.253.150:9000'`

## Setup the chain

First get the address of all nodes by running:

```sh
sudo memcoin --config /var/opt/dedis/dvoting/data/dela ordering export
```

This will yield a base64 encoded string `<ADDRESS>:<PUB KEY>`.

From the first node.

**1: Create the chain**:

Include ALL nodes, the first and all other nodes.

```sh
sudo memcoin --config /var/opt/dedis/dvoting/data/dela ordering setup \
    --member <RESULT FROM ordering export>\
    --member <...>
    ...
```

**2: grant access for each node to sign transactions on the evoting smart contract**:

To be done on each node.

```sh
PK=<> # taken from the "ordering export", the part after ":"
sudo memcoin --config /var/opt/dedis/dvoting/data/dela pool add \
    --key $keypath \
    --args go.dedis.ch/dela.ContractArg --args go.dedis.ch/dela.Access \
    --args access:grant_id --args 0300000000000000000000000000000000000000000000000000000000000000 \
    --args access:grant_contract --args go.dedis.ch/dela.Evoting \
    --args access:grant_command --args all \
    --args access:identity --args $PK \
    --args access:command --args GRANT
```

# Package D-Voting in an installable .deb file

A .deb package is created by the CI upon the creation of a release. You might
want to check `deb-package/upload-artifacts.sh` and the `go_release.yml` action.

## Requirements

- gem
- build-essential
- git
- fpm (`sudo gem install fpm`)
- go (see https://go.dev/doc/install)

```sh
sudo apt install rubygems build-essential git
```

## Get the code

```sh
git clone https://github.com/dedis/d-voting.git 
```

## Build the deb package

from the root folder, use make:

```sh
make deb
```

Make sure that a git tag exist, i.e `git describe` shows your tag.

The resulting .deb can be found in the `dist/` folder.

## Debian deployment

A package registry with debian packages is available at http://apt.dedis.ch.
To install a package run the following:

```sh
echo "deb http://apt.dedis.ch/ squeeze main" >> /etc/apt/sources.list
wget -q -O- http://apt.dedis.ch/dvoting-release.pgp | sudo apt-key add -
sudo apt update
sudo apt install dedis-dvoting
```