# eth-statediff-service

[![Go Report Card](https://goreportcard.com/badge/github.com/vulcanize/eth-statediff-service)](https://goreportcard.com/report/github.com/vulcanize/eth-statediff-service)

>> standalone statediffing service ontop of leveldb

Purpose:

Stand up a statediffing service directly on top of a go-ethereum leveldb instance.
This service can serve historical state data over the same rpc interface as
[statediffing geth](https://github.com/vulcanize/go-ethereum/releases/tag/v1.9.11-statediff-0.0.5) without needing to run a full node

## Usage

### `serve`

To serve state diffs over RPC:

`eth-statediff-service serve --config=<config path>`

Available RPC methods are:
  * `statediff_stateTrieAt()`
  * `statediff_streamCodeAndCodeHash()`
  * `statediff_stateDiffAt()`
  * `statediff_writeStateDiffAt()`

### `write`

To write state diffs directly to a database:

`eth-statediff-service write --config=<config path>`

This depends on the `database` settings being properly configured.

## Configuration

An example config file:

```toml
[leveldb]
    path = "/Users/user/Library/Ethereum/geth/chaindata"
    ancient = "/Users/user/Library/Ethereum/geth/chaindata/ancient"

[server]
    ipcPath = "~/.vulcanize/vulcanize.ipc"
    httpPath = "127.0.0.1:8545"

[write]
    ranges = [[1, 2], [3, 4]]
[write.params]
    IntermediateStateNodes   = true
    IntermediateStorageNodes = false
    IncludeBlock             = true
    IncludeReceipts          = true
    IncludeTD                = true
    IncludeCode              = false

[statediff]
    workers = 4

[log]
    file = "~/.vulcanize/statediff.log"
    level = "info"

[eth]
    chainID = 1

```
