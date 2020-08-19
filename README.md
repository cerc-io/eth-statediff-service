## eth-statediff-service
>> standalone statediffing service ontop of leveldb

Purpose:

Stand up a statediffing service directly on top of a go-ethereum leveldb instance.
This service can serve historical state data over the same rpc interface as
[statediffing geth](https://github.com/vulcanize/go-ethereum/releases/tag/v1.9.11-statediff-0.0.5) without needing to run a full node

Usage:

`./eth-statediff-service serve --config={path to toml config file}`

Config:

```toml
[leveldb]
    path = "/Users/user/Library/Ethereum/geth/chaindata"
    ancient = "/Users/user/Library/Ethereum/geth/chaindata/ancient"

[server]
    ipcPath = "~/.vulcanize/vulcanize.ipc"
    httpPath = "127.0.0.1:8545"

[log]
    file = ""
    level = "info"

[eth]
    chainID = 1
```
