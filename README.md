# eth-statediff-service

[![Go Report Card](https://goreportcard.com/badge/github.com/vulcanize/eth-statediff-service)](https://goreportcard.com/report/github.com/vulcanize/eth-statediff-service)

>> standalone statediffing service on top of LevelDB

Purpose:

Stand up a statediffing service directly on top of a go-ethereum LevelDB instance.
This service can serve historical state data over the same rpc interface as
[statediffing geth](https://github.com/vulcanize/go-ethereum) without needing to run a full node.

## Setup

Build the binary:

```bash
make build
```

## Configuration

An example config file:

```toml
[leveldb]
    # LevelDB access mode <local | remote>
    mode = "local"  # LVLDB_MODE

    # in local mode
    # LevelDB paths
    path    = "/Users/user/Library/Ethereum/geth/chaindata"         # LVLDB_PATH
    ancient = "/Users/user/Library/Ethereum/geth/chaindata/ancient" # LVLDB_ANCIENT

    # in remote mode
    # URL for leveldb-ethdb-rpc endpoint
    url = "http://127.0.0.1:8082/"  # LVLDB_URL

[server]
    ipcPath  = ".ipc"           # SERVICE_IPC_PATH
    httpPath = "127.0.0.1:8545" # SERVICE_HTTP_PATH

[statediff]
    prerun          = true  # STATEDIFF_PRERUN
    serviceWorkers  = 1     # STATEDIFF_SERVICE_WORKERS
    workerQueueSize = 1024  # STATEDIFF_WORKER_QUEUE_SIZE
    trieWorkers     = 4     # STATEDIFF_TRIE_WORKERS

[prerun]
    only = false    # PRERUN_ONLY

    # to perform prerun in a specific range (optional)
    start = 0   # PRERUN_RANGE_START
    stop  = 100 # PRERUN_RANGE_STOP

    # to perform prerun over multiple ranges (optional)
    ranges = [
        [101, 1000]
    ]

    # statediffing params for prerun
    [prerun.params]
        intermediateStateNodes   = true # PRERUN_INTERMEDIATE_STATE_NODES
        intermediateStorageNodes = true # PRERUN_INTERMEDIATE_STORAGE_NODES
        includeBlock             = true # PRERUN_INCLUDE_BLOCK
        includeReceipts          = true # PRERUN_INCLUDE_RECEIPTS
        includeTD                = true # PRERUN_INCLUDE_TD
        includeCode              = true # PRERUN_INCLUDE_CODE
        watchedAddresses         = []

[log]
    file  = ""      # LOG_FILE_PATH
    level = "info"  # LOG_LEVEL

[database]
    # output type <postgres | file | dump>
    type = "postgres"

    # with postgres type
    # db credentials
    name     = "vulcanize_test" # DATABASE_NAME
    hostname = "localhost"      # DATABASE_HOSTNAME
    port     = 5432             # DATABASE_PORT
    user     = "vulcanize"      # DATABASE_USER
    password = "..."            # DATABASE_PASSWORD
    driver   = "sqlx"           # DATABASE_DRIVER_TYPE <sqlx | pgx>

    # with file type
    # file mode <sql | csv>
    fileMode = "csv"    # DATABASE_FILE_MODE

    # with SQL file mode
    filePath = ""   # DATABASE_FILE_PATH

    # with CSV file mode
    fileCsvDir = "output_dir" # DATABASE_FILE_CSV_DIR

    # with dump type
    # <stdout | stderr | discard>
    dumpDestination = ""    # DATABASE_DUMP_DST

[cache]
    database = 1024 # DB_CACHE_SIZE_MB
    trie     = 1024 # TRIE_CACHE_SIZE_MB

[prom]
    # prometheus metrics
    metrics  = true         # PROM_METRICS
    http     = true         # PROM_HTTP
    httpAddr = "localhost"  # PROM_HTTP_ADDR
    httpPort = "8889"       # PROM_HTTP_PORT

[ethereum]
    # node info
    nodeID       = ""                       # ETH_NODE_ID
    clientName   = "eth-statediff-service"  # ETH_CLIENT_NAME
    genesisBlock = "0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3" # ETH_GENESIS_BLOCK
    networkID    = 1                        # ETH_NETWORK_ID
    chainID      = 1                        # ETH_CHAIN_ID

    # path to custom chain config file (optional)
    # keep chainID same as that in chain config file
    chainConfig  = "./chain.json"           # ETH_CHAIN_CONFIG
```

### Local Setup

* Create a chain config file `chain.json` according to chain config in genesis json file used by local geth.

  Example:
  ```json
  {
    "chainId": 41337,
    "homesteadBlock": 0,
    "eip150Block": 0,
    "eip150Hash": "0x0000000000000000000000000000000000000000000000000000000000000000",
    "eip155Block": 0,
    "eip158Block": 0,
    "byzantiumBlock": 0,
    "constantinopleBlock": 0,
    "petersburgBlock": 0,
    "istanbulBlock": 0,
    "clique": {
      "period": 5,
      "epoch": 30000
    }
  }
  ```

  Provide the path to the above file in the config.

## Usage

* Create / update the config file (refer to example config above).

### `serve`

* To serve state diffs over RPC:

    ```bash
    ./eth-statediff-service serve --config=<config path>
    ```

    Example:

    ```bash
    ./eth-statediff-service serve --config environments/config.toml
    ```

* Available RPC methods are:
    * `statediff_stateTrieAt()`
    * `statediff_streamCodeAndCodeHash()`
    * `statediff_stateDiffAt()`
    * `statediff_writeStateDiffAt()`
    * `statediff_writeStateDiffsInRange()`

    Example:

    ```bash
    curl -X POST -H 'Content-Type: application/json' --data '{"jsonrpc":"2.0","method":"statediff_writeStateDiffsInRange","params":['"$BEGIN"', '"$END"', {"intermediateStateNodes":true,"intermediateStorageNodes":true,"includeBlock":true,"includeReceipts":true,"includeTD":true,"includeCode":true}],"id":1}' "$HOST":"$PORT"
    ```

* Prerun:
    * The process can be configured locally with sets of ranges to process as a "prerun" to processing directed by the server endpoints.
    * This is done by turning "prerun" on in the config (`statediff.prerun = true`) and defining ranges and params in the
`prerun` section of the config.
    * Set the range using `prerun.start` and `prerun.stop`. Use `prerun.ranges` if prerun on more than one range is required.
    * Currently, `prerun.params.includeTD` must be set to `true`.

## Import output data in file mode into a database

* When `eth-statediff-service` is run in file mode (`database.type`) the output is in form of a SQL file or multiple CSV files.

* Assuming the output files are located in `./output_dir` directory, if the DB is running in docker we need to mount the directory containing the files as a volume in the DB service. Eg:

    ```yaml
    # In docker-compose file
    services:
      ipld-eth-db:
        volumes:
          - ./output_dir:/output_dir
    ```

* Start `bash` in the DB container to run import commands:

    ```bash
    docker exec -it <CONTAINER_ID> bash
    ```

### SQL

* De-duplicate data:

    ```bash
    sort -u output_dir/statediff.sql -o output_dir/statediff.sql
    ```

* Run the following to import data:

    ```bash
    psql -U <DATABASE_USER> <DATABASE_NAME> --set ON_ERROR_STOP=on -f /output_dir/statediff.sql
    ```

### CSV

- De-duplicate data:

    ```bash
    # public.blocks
    sort -u output_dir/public.blocks.csv -o output_dir/public.blocks.csv
    ```

* Run `psql`:

    ```bash
    psql -U <DATABASE_USER> <DATABASE_NAME>
    ```

* Run the following to import data:

    ```bash
    # public.nodes
    COPY public.nodes FROM '/output_dir/public.nodes.csv' CSV;

    # public.nodes
    COPY public.blocks FROM '/output_dir/public.blocks.csv' CSV;

    # eth.header_cids
    COPY eth.header_cids FROM '/output_dir/eth.header_cids.csv' CSV;

    # eth.uncle_cids
    COPY eth.uncle_cids FROM '/output_dir/eth.uncle_cids.csv' CSV;

    # eth.transaction_cids
    COPY eth.transaction_cids FROM '/output_dir/eth.transaction_cids.csv' CSV FORCE NOT NULL dst;

    # eth.access_list_elements
    COPY eth.access_list_elements FROM '/output_dir/eth.access_list_elements.csv' CSV;

    # eth.receipt_cids
    COPY eth.receipt_cids FROM '/output_dir/eth.receipt_cids.csv' CSV FORCE NOT NULL post_state, contract, contract_hash;

    # eth.log_cids
    COPY eth.log_cids FROM '/output_dir/eth.log_cids.csv' CSV FORCE NOT NULL topic0, topic1, topic2, topic3;

    # eth.state_cids
    COPY eth.state_cids FROM '/output_dir/eth.state_cids.csv' CSV FORCE NOT NULL state_leaf_key;

    # eth.storage_cids
    COPY eth.storage_cids FROM '/output_dir/eth.storage_cids.csv' CSV FORCE NOT NULL storage_leaf_key;

    # eth.state_accounts
    COPY eth.state_accounts FROM '/output_dir/eth.state_accounts.csv' CSV;
    ```