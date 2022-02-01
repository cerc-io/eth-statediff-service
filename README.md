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
  * `statediff_writeStateDiffsInRange()`

e.g. `curl -X POST -H 'Content-Type: application/json' --data '{"jsonrpc":"2.0","method":"statediff_writeStateDiffsInRange","params":['"$BEGIN"', '"$END"', {"intermediateStateNodes":true,"intermediateStorageNodes":true,"includeBlock":true,"includeReceipts":true,"includeTD":true,"includeCode":true}],"id":1}' "$HOST":"$PORT"`

The process can be configured locally with sets of ranges to process as a "prerun" to processing directed by the server endpoints.
This is done by turning "prerun" on in the config (`statediff.prerun = true`) and defining ranged and params in the
`prerun` section of the config as shown below.

## Configuration

An example config file:

```toml
[leveldb]
    path = "/Users/user/Library/Ethereum/geth/chaindata"
    ancient = "/Users/user/Library/Ethereum/geth/chaindata/ancient"

[server]
    ipcPath = ".ipc"
    httpPath = "127.0.0.1:8545"

[statediff]
    prerun = true
    serviceWorkers = 1
    workerQueueSize = 1024
    trieWorkers = 4

[prerun]
    only = false
    ranges = [
        [0, 1000]
    ]
    [prerun.params]
        intermediateStateNodes = true
        intermediateStorageNodes = true
        includeBlock = true
        includeReceipts = true
        includeTD = true
        includeCode = true
        watchedAddresses = []
        watchedStorageKeys = []

[log]
    file = ""
    level = "info"

[eth]
    chainID = 1

[database]
    name     = "vulcanize_test"
    hostname = "localhost"
    port     = 5432
    user     = "vulcanize"
    password = "..."
    type = "postgres"
    driver = "sqlx"
    dumpDestination = ""
    filePath = ""

[cache]
    database = 1024
    trie = 1024

[prom]
    dbStats = false
    metrics = true
    http = true
    httpAddr = "localhost"
    httpPort = "8889"

[ethereum]
    nodeID = ""
    clientName = "eth-statediff-service"
    genesisBlock = "0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3"
    networkID = 1
    chainID = 1
```
