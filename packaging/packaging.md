# Packaging D-Voting in an installable .deb file

## Requirements

- Install `gem` and [fpm](https://fpm.readthedocs.io/en/latest/installation.html):

```sh
sudo apt install rubygems build-essentials
sudo gem install fpm
```

- Build memcoin:

```sh
cd cli/memcoin
go build
```

- Build unikernel. See [how to build unikernel](../contracts/evoting/unikernel/apps/combine_shares/README.md):

```sh
cd contracts/evoting/unikernel/apps/combine_shares
git submodule update --init --recursive 
make menuconfig
make
```

- Create .deb package:

```sh
cd packaging
./build-deb.sh
```
