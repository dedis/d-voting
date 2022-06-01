# Packaging D-Voting in an installable .deb file

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
git clone --branch packaging https://github.com/dedis/d-voting.git --recursive 
```

## Build the deb package

from the root folder, use make:

```sh
make deb
```

Make sure that a git tag exist, i.e `git describe` shows your tag.

The resulting .deb can be found in the `dist/` folder.

## Things to do after install

### Network config

Ensure that the public address is correct. For instance, in `network.env`, replace:
```sh
export dela_public="//localhost:9000"
```
with the node's public address:
```sh
export dela_public="//172.16.253.150:9000"
```

### Leader's node

Get the token and certificate (24h * 30 = 720):

```sh
sudo memcoin --config /var/opt/dedis/dvoting/data/dela minogrpc token \
    --expiration 720h
```

This result, which looks like as follow, will be given to node's operators:

```
--token b6VhdQEPXKOtZHpng8E8jw== --cert-hash oNeyrA864P2cP+TT6IE6GvkeEI/Ec4rOlZWEWiQkQKk=
```

### Participants (node's operators)

Join the network. This operation will make the node share its certificate to the
MASTER node, which, in turn, will share its known certificates to the node. Note
that the certificates are stored in the DB, which means that this operation must
be re-done in case the DB is reset.

```sh
sudo memcoin --config /var/opt/dedis/dvoting/data/dela minogrpc join \
    --address <MASTER NODE ADDRESS> --token <TOKEN> --cert-hash <CERT HASH>
```

Example of `<MASTER NODE ADDRESS>`: `'//172.16.253.150:9000'`

Get the node's address and public key:

```sh
sudo memcoin --config /var/opt/dedis/dvoting/data/dela ordering export
```

This will yield a base64 encoded string `<ADDRESS>:<PUB KEY>`.

It will have to be provided to EPFL.

## Setup the chain, from EPFL

**1: Create the chain**:

Do not forget to include ourself, the EPFL node!

```sh
sudo memcoin --config /var/opt/dedis/dvoting/data/dela ordering setup \
    --member <RESULT FROM ordering export>\
    --member <...>
    ...
```

**2: grant access for each node to sign transactions on the evoting smart contract**:

```sh
PK=<> # taken from the "ordering export", the part after ":"
sudo memcoin --config /var/opt/dedis/dvoting/data/dela pool add \
    --key /home/user/master.key \
    --args go.dedis.ch/dela.ContractArg --args go.dedis.ch/dela.Access \
    --args access:grant_id --args 0300000000000000000000000000000000000000000000000000000000000000 \
    --args access:grant_contract --args go.dedis.ch/dela.Evoting \
    --args access:grant_command --args all \
    --args access:identity --args $PK \
    --args access:command --args GRANT
```

You should also grant access to the master key.

### Test

```sh
sudo memcoin --config /var/opt/dedis/dvoting/data/dela e-voting scenarioTest \
    --proxy-addr1 "http://192.168.232.133:9080" \
    --proxy-addr2 "http://192.168.232.134:9080" \
    --proxy-addr3 "http://192.168.232.135:9080"
```
