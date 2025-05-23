## Go Zond

Official Golang execution layer implementation of the Zond protocol.

[![Go Report Card](https://goreportcard.com/badge/github.com/theQRL/go-zond)](https://goreportcard.com/report/github.com/theQRL/go-zond)
[![Discord](https://img.shields.io/badge/discord-join%20chat-blue.svg)](https://www.theqrl.org/discord)

**This code is a test release. All code, features and documentation are subject to change and may represent a work in progress**

## Prerequisites

- Go version 1.21 or later (Note: Go 1.23.6 has known compatibility issues on Windows - please use Go 1.22.12 or earlier for Windows builds)
- Git
- Make (optional, for using Makefile commands)

## Building the source

For prerequisites and detailed build instructions please read the [Installation Instructions](https://test-zond.theqrl.org/install).

Building `gzond` requires both a Go (version 1.21 or later) and a C compiler. You can install
them using your favourite package manager. Once the dependencies are installed, run

```shell
make gzond
```

or, to build the full suite of utilities:

```shell
make all
```

## Executables

The go-zond project comes with several wrappers/executables found in the `cmd`
directory.

|  Command    | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
| :--------:  | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **`gzond`** | Our main Zond CLI client. It is the entry point into the Zond network (main-, test- or private net), capable of running as a full node (default), archive node (retaining all historical state) or a light node (retrieving data live). It can be used by other processes as a gateway into the Zond network via JSON RPC endpoints exposed on top of HTTP, WebSocket and/or IPC transports. Based on geth, `gzond --help` and the [geth CLI page](https://geth.ethereum.org/docs/fundamentals/command-line-options) show command line options. |
|   `clef`    | Stand-alone signing tool, which can be used as a backend signer for `gzond`.                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
|  `devp2p`   | Utilities to interact with nodes on the networking layer, without running a full blockchain.                                                                                                                                                                                                                                                                                                                                                                                                                                       |
|  `abigen`   | Source code generator to convert Zond contract definitions into easy-to-use, compile-time type-safe Go packages. It operates on plain [Zond contract ABIs](https://docs.soliditylang.org/en/develop/abi-spec.html) with expanded functionality if the contract bytecode is also available. However, it also accepts Hyperion source files, making development much more streamlined. Please see the [Native DApps](https://geth.ethereum.org/docs/developers/dapp-developer/native-bindings) page for details.                                  |
| `bootnode`  | Stripped down version of our Zond client implementation that only takes part in the network node discovery protocol, but does not run any of the higher level application protocols. It can be used as a lightweight bootstrap node to aid in finding peers in private networks.                                                                                                                                                                                                                                               |
|   `zvm`     | Developer utility version of the ZVM (Zond Virtual Machine) that is capable of running bytecode snippets within a configurable environment and execution mode. Its purpose is to allow isolated, fine-grained debugging of ZVM opcodes (e.g. `zvm --code 60ff60ff --debug run`).                                                                                                                                                                                                                                               |
| `rlpdump`   | Developer utility tool to convert binary RLP ([Recursive Length Prefix](https://ethereum.org/en/developers/docs/data-structures-and-encoding/rlp)) dumps (data encoding used by the Zond protocol both network as well as consensus wise) to user-friendlier hierarchical representation (e.g. `rlpdump --hex CE0183FFFFFFC4C304050583616263`).                                                                                                                                                                                |

## Running `gzond`

Going through all the possible command line flags is out of scope here (please see our nascent [Zond Testnet docs](https://test-zond.theqrl.org) or consult the
[geth CLI Wiki page](https://geth.ethereum.org/docs/fundamentals/command-line-options)),
but we've enumerated a few common parameter combos to get you up to speed quickly
on how you can run your own `gzond` instance.

### Hardware Requirements

Minimum:

* CPU with 2+ cores
* 4GB RAM
* 1TB free storage space to sync the Mainnet
* 8 MBit/sec download Internet service

Recommended:

* Fast CPU with 4+ cores
* 16GB+ RAM
* High-performance SSD with at least 1TB of free space
* 25+ MBit/sec download Internet service

### Full node on the main Zond network

By far the most common scenario is people wanting to simply interact with the Zond
network: create accounts; transfer funds; deploy and interact with contracts. For this
particular use case, the user doesn't care about years-old historical data, so we can
sync quickly to the current state of the network. To do so:

```shell
$ gzond console
```

This command will:
 * Start `gzond` in snap sync mode (default, can be changed with the `--syncmode` flag),
   causing it to download more data in exchange for avoiding processing the entire history
   of the Zond network, which is very CPU intensive.
 * Start the built-in interactive [JavaScript console](https://geth.ethereum.org/docs/interacting-with-geth/javascript-console),
   (via the trailing `console` subcommand) through which you can interact using [`web3` methods](https://github.com/ChainSafe/web3.js/blob/0.20.7/DOCUMENTATION.md) 
   (note: the `web3` version bundled within `gzond` is very old, and not up to date with official docs),
   as well as `gzond`'s own [management APIs](https://geth.ethereum.org/docs/interacting-with-geth/rpc).
   This tool is optional and if you leave it out you can always attach it to an already running
   `gzond` instance with `gzond attach`.

### Configuration

As an alternative to passing the numerous flags to the `gzond` binary, you can also pass a
configuration file via:

```shell
$ gzond --config /path/to/your_config.toml
```

To get an idea of how the file should look like you can use the `dumpconfig` subcommand to
export your existing configuration:

```shell
$ gzond --your-favourite-flags dumpconfig
```

#### Docker quick start

_Docker deployment in development_

One of the quickest ways to get Zond up and running on your machine is by using
Docker:

```shell
docker run -d --name zond-node -v /Users/alice/zond:/root \
           -p 8545:8545 -p 30303:30303 \
           theqrl/gzond
```

This will start `gzond` in snap-sync mode with a DB memory allowance of 1GB, as the
above command does.  It will also create a persistent volume in your home directory for
saving your blockchain as well as map the default ports. There is also an `alpine` tag
available for a slim version of the image.

Do not forget `--http.addr 0.0.0.0`, if you want to access RPC from other containers
and/or hosts. By default, `gzond` binds to the local interface and RPC endpoints are not
accessible from the outside.

### Programmatically interfacing `gzond` nodes

As a developer, sooner rather than later you'll want to start interacting with `gzond` and the
Zond network via your own programs and not manually through the console. To aid
this, `gzond` has built-in support for Ethereum-compatible, JSON-RPC based APIs ([standard APIs](https://ethereum.github.io/execution-apis/api-documentation/)
and [`gzond` specific APIs](https://geth.ethereum.org/docs/interacting-with-geth/rpc)).
These can be exposed via HTTP, WebSockets and IPC (UNIX sockets on UNIX based
platforms, and named pipes on Windows).

The IPC interface is enabled by default and exposes all the APIs supported by `gzond`,
whereas the HTTP and WS interfaces need to manually be enabled and only expose a
subset of APIs due to security reasons. These can be turned on/off and configured as
you'd expect.

HTTP based JSON-RPC API options:

  * `--http` Enable the HTTP-RPC server
  * `--http.addr` HTTP-RPC server listening interface (default: `localhost`)
  * `--http.port` HTTP-RPC server listening port (default: `8545`)
  * `--http.api` API's offered over the HTTP-RPC interface (default: `zond,net,web3`)
  * `--http.corsdomain` Comma separated list of domains from which to accept cross origin requests (browser enforced)
  * `--ws` Enable the WS-RPC server
  * `--ws.addr` WS-RPC server listening interface (default: `localhost`)
  * `--ws.port` WS-RPC server listening port (default: `8546`)
  * `--ws.api` API's offered over the WS-RPC interface (default: `zond,net,web3`)
  * `--ws.origins` Origins from which to accept WebSocket requests
  * `--ipcdisable` Disable the IPC-RPC server
  * `--ipcapi` API's offered over the IPC-RPC interface (default: `admin,debug,zond,miner,net,personal,txpool,web3`)
  * `--ipcpath` Filename for IPC socket/pipe within the datadir (explicit paths escape it)

You'll need to use your own programming environments' capabilities (libraries, tools, etc) to
connect via HTTP, WS or IPC to a `gzond` node configured with the above flags and you'll
need to speak [JSON-RPC](https://www.jsonrpc.org/specification) on all transports. You
can reuse the same connection for multiple requests!

**Note: Please understand the security implications of opening up an HTTP/WS based
transport before doing so! Hackers on the internet are actively trying to subvert
Zond nodes with exposed APIs! Further, all browser tabs can access locally
running web servers, so malicious web pages could try to subvert locally available
APIs!**

### Operating a private network

Maintaining your own private network is more involved as a lot of configurations taken for
granted in the official networks need to be manually set up.

#### Defining the private genesis state

First, you'll need to create the genesis state of your networks, which all nodes need to be
aware of and agree upon. This consists of a small JSON file (e.g. call it `genesis.json`):

```json
{
  "config": {
    "chainId": <arbitrary positive integer>
  },
  "alloc": {},
  "coinbase": "Z0000000000000000000000000000000000000000",
  "extraData": "",
  "gasLimit": "0x2fefd8",
  "mixhash": "0x0000000000000000000000000000000000000000000000000000000000000000",
  "parentHash": "0x0000000000000000000000000000000000000000000000000000000000000000",
  "timestamp": "0x00"
}
```

The above fields should be fine for most purposes, although we'd recommend changing
the `nonce` to some random value so you prevent unknown remote nodes from being able
to connect to you. If you'd like to pre-fund some accounts for easier testing, create
the accounts and populate the `alloc` field with their addresses.

```json
"alloc": {
  "Z0000000000000000000000000000000000000001": {
    "balance": "111111111"
  },
  "Z0000000000000000000000000000000000000002": {
    "balance": "222222222"
  }
}
```

With the genesis state defined in the above JSON file, you'll need to initialize **every**
`gzond` node with it prior to starting it up to ensure all blockchain parameters are correctly
set:

```shell
$ gzond init path/to/genesis.json
```

#### Creating the rendezvous point

With all nodes that you want to run initialized to the desired genesis state, you'll need to
start a bootstrap node that others can use to find each other in your network and/or over
the internet. The clean way is to configure and run a dedicated bootnode:

```shell
$ bootnode --genkey=boot.key
$ bootnode --nodekey=boot.key
```

With the bootnode online, it will display an [`enode` URL](https://ethereum.org/en/developers/docs/networking-layer/network-addresses/#enode)
that other nodes can use to connect to it and exchange peer information. Make sure to
replace the displayed IP address information (most probably `[::]`) with your externally
accessible IP to get the actual `enode` URL.

*Note: You could also use a full-fledged `gzond` node as a bootnode, but it's the less
recommended way.*

#### Starting up your member nodes

With the bootnode operational and externally reachable (you can try
`telnet <ip> <port>` to ensure it's indeed reachable), start every subsequent `gzond`
node pointed to the bootnode for peer discovery via the `--bootnodes` flag. It will
probably also be desirable to keep the data directory of your private network separated, so
do also specify a custom `--datadir` flag.

```shell
$ gzond --datadir=path/to/custom/data/folder --bootnodes=<bootnode-enode-url-from-above>
```

*Note: Since your network will be completely cut off from the main and test networks, you'll
also need to configure a miner to process transactions and create new blocks for you.*

## Contribution

Thank you for considering helping out with the source code! We welcome contributions
from anyone on the internet, and are grateful for even the smallest of fixes!

If you'd like to contribute to go-zond, please fork, fix, commit and send a pull request
for the maintainers to review and merge into the main code base. If you wish to submit
more complex changes though, please check up with the core devs first on [our Discord Server](https://theqrl.org/discord)
to ensure those changes are in line with the general philosophy of the project and/or get
some early feedback which can make both your efforts much lighter as well as our review
and merge procedures quick and simple.

Please make sure your contributions adhere to our coding guidelines:

* Code must adhere to the official Go [formatting](https://golang.org/doc/effective_go.html#formatting)
   guidelines (i.e. uses [gofmt](https://golang.org/cmd/gofmt/)).
* Code must be documented adhering to the official Go [commentary](https://golang.org/doc/effective_go.html#commentary)
   guidelines.
* Pull requests need to be based on and opened against the `main` branch.
* Commit messages should be prefixed with the package(s) they modify.
* E.g. "zond, rpc: make trace configs optional"

## License

The go-zond library (i.e. all code outside of the `cmd` directory) is licensed under the
[GNU Lesser General Public License v3.0](https://www.gnu.org/licenses/lgpl-3.0.en.html),
also included in our repository in the `COPYING.LESSER` file.

The go-zond binaries (i.e. all code inside of the `cmd` directory) are licensed under the
[GNU General Public License v3.0](https://www.gnu.org/licenses/gpl-3.0.en.html), also
included in our repository in the `COPYING` file.
