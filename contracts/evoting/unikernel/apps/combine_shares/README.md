# Proof-of-Concept of Configurable Smart Contract in C

Credits to: https://github.com/cs-pub-ro/app-smart-contract-config

This is a proof-of-concept smart contract program running on Unikraft. It uses
command line arguments, TCP networking and 9pfs filesystem mounting to configure
the behavior of the smart contract. It exposes a basic cryptographic hash
functionality provided by [libsodium](https://github.com/jedisct1/libsodium).

## Requirements

```sh
sudo apt-get install git
sudo apt-get install make
sudo apt-get install gcc
sudo apt-get install libncurses-dev
sudo apt-get install flex
sudo apt-get install bison
sudo apt-get install unzip
sudo apt-get install socat
sudo apt-get install uuid-runtime
sudo apt install qemu-kvm libvirt-daemon-system libvirt-clients bridge-utils \
    virtinst virt-manager
sudo apt install net-tools
```

## Setup

It is easiest to use a file hierarchy such as:

```tree
unikernel
 ┣ apps
 ┃ ┗ combine_shares
 ┣ libs
 ┃ ┣ libsodium
 ┃ ┣ lwip
 ┃ ┗ newlib
 ┗ unikraft
```

Please run the following command, which will clone the submodules in the above 
illustrated paths:

```sh
git submodule update --init --recursive
```

If you use a file hierarchy different than the one above, then edit the
`Makefile` file and update the variables to point to the correct location:

* Update the `UK_ROOT` variable to point to the location of the Unikraft clone.
* Update the `UK_LIBS` variable to point to the folder storing the library
  clones.

If you get a `Cannot find library` error at one of the following steps, you may
need to add the `lib-` prefix to the libraries listed in the `LIBS` variable.
For instance, `$(UK_LIBS)/lwip` becomes `$(UK_LIBS)/lib-lwip`.

Here is the list of the required repositories that you also need as submodules:

* [libsodium](https://github.com/unikraft/lib-libsodium)
* [lwip](https://github.com/unikraft/lib-lwip)
* [newlib](https://github.com/unikraft/lib-newlib)
* [Unikraft](https://github.com/unikraft/unikraft)

For the libsodium library, the submodule still points to the source of the following PR:
https://github.com/unikraft/lib-libsodium/pull/4.

For the other libraries, please make sure that each repository is on the `staging` branch. 


## Configure

Configure the build by running in `contracts/evoting/unikernel/apps/combine_shares`:

```sh
make menuconfig
```

The basic configuration is loaded from the `Config.uk` file. The basic
configuration is loaded from the `Config.uk` file. Then add support for 9pfs, by
selecting `Library Configuration` -> `vfscore: VFS Core Interface` -> `vfscore:
VFS Configuration`. The select `Automatically mount a root filesystem (/)` and
select `9pfs`. For the `Default root device` option fill `fs0`. You should end
up with an image such as the one in `9pfs_config.png`. Save the configuration
and exit.

As a result of this, a `.config` file is created.
A KVM unikernel image will be built from the configuration.

## Build

Build the unikernel KVM image by running:

```sh
make
```

The resulting image is in the `build/` subfolder, usually named
`build/combine_shares_kvm-x86_64`.

## Run

The resulting image is to be loaded as part of a KVM virtual machine with a
software bridge allowing outside network access to the server application. Start
the image by using the `run` script (you require `sudo` privileges):

```sh
./run
```

```
[...]
Booting from ROM...
1: Set IPv4 address 172.44.0.2 mask 255.255.255.0 gw 172.44.0.1
en1: Added
en1: Interface is up
Powered by
o.   .o       _ _               __ _
Oo   Oo  ___ (_) | __ __  __ _ ' _) :_
oO   oO ' _ `| | |/ /  _)' _` | |_|  _)
oOo oOO| | | | |   (| | | (_) |  _) :_
 OoOoO ._, ._:_:_,\_._,  .__,_:_, \___)
                    Dione 0.6.0~6a2069e
argument is 1024
Listening on port 1024...
```

The `1024` is passed as a command line argument.

The server now listens for "commands". These commands are files in the `mnt/`
folder that will be read and processed by the smart contract program. Passing
the commands `ec_multiply` will result in the program reading the file from
the `mnt/` folder.

On another terminal, use `nc` to connect via TCP to the program and send the
commands:

```sh
nc 172.44.0.2 1024
ec_multiply
2c7be86ab07488ba43e8e03d85a67625cfbf98c8544de4c877241b7aaafc7fe3
```

The running program will print out a summary of received commands:

```sh
Received connection from 172.44.0.1:38596
Received: ec_multiply
Operation: ec_multiply
Sent: 2c7be86ab07488ba43e8e03d85a67625cfbf98c8544de4c877241b7aaafc7fe3
```

When done, close the server / KVM machine using `Ctrl+c`.
