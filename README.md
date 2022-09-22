# eth-statediff-service

[![Go Report Card](https://goreportcard.com/badge/github.com/vulcanize/eth-statediff-service)](https://goreportcard.com/report/github.com/vulcanize/eth-statediff-service)

>> standalone statediffing service on top of LevelDB

Purpose:

Stand up a statediffing service directly on top of a go-ethereum LevelDB instance.
This service can serve historical state data over the same rpc interface as
[statediffing geth](https://github.com/cerc-io/go-ethereum) without needing to run a full node.

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
    only = false     # PRERUN_ONLY
    parallel = true  # PRERUN_PARALLEL

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
    dbStats = true          # PROM_DB_STATS

[ethereum]
    # node info
    nodeID       = ""                       # ETH_NODE_ID
    clientName   = "eth-statediff-service"  # ETH_CLIENT_NAME
    networkID    = 1                        # ETH_NETWORK_ID
    chainID      = 1                        # ETH_CHAIN_ID
    genesisBlock = "0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3" # ETH_GENESIS_BLOCK

    # path to custom chain config file (optional)
    # keep chainID same as that in chain config file
    chainConfig  = "./chain.json"           # ETH_CHAIN_CONFIG

[debug]
    pprof = false                           # DEBUG_PPROF
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

* To serve the statediff RPC API:

    ```bash
    ./eth-statediff-service serve --config=<config path>
    ```

    Example:

    ```bash
    ./eth-statediff-service serve --config environments/config.toml
    ```

* Available RPC methods:
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

* NOTE: Currently, `params.includeTD` must be set to / passed as `true`.

## Monitoring

* Enable metrics using config parameters `prom.metrics` and `prom.http`.
* `eth-statediff-service` exposes following prometheus metrics at `/metrics` endpoint:
    * `ranges_queued`: Number of range requests currently queued.
    * `loaded_height`: The last block that was loaded for processing.
    * `processed_height`: The last block that was processed.
    * `stats.t_block_load`: Block loading time.
    * `stats.t_block_processing`: Block (header, uncles, txs, rcts, tx trie, rct trie) processing time.
    * `stats.t_state_processing`: State (state trie, storage tries, and code) processing time.
    * `stats.t_postgres_tx_commit`: Postgres tx commit time.
    * `http.count`: HTTP request count.
    * `http.duration`: HTTP request duration.
    * `ipc.count`: Unix socket connection count.

## Tests

* Run unit tests:

    ```bash
    make test
    ```

## Import output data in file mode into a database

* When `eth-statediff-service` is run in file mode (`database.type`) the output is in form of a SQL file or multiple CSV files.

### SQL

* Assuming the output files are located in host's `./output_dir` directory.

* Create a directory to store post-processed output:

    ```bash
    mkdir -p output_dir/processed_output
    ```

* (Optional) Get row counts in the output:

    ```bash
    wc -l output_dir/statediff.sql > output_stats.txt
    ```

* De-duplicate data:

    ```bash
    sort -u output_dir/statediff.sql -o output_dir/processed_output/deduped-statediff.sql
    ```

* Copy over the post-processed output files to the DB server (say in `/output_dir`).

* Run the following to import data:

    ```bash
    psql -U <DATABASE_USER> -h <DATABASE_HOSTNAME> -p <DATABASE_PORT> <DATABASE_NAME> --set ON_ERROR_STOP=on -f /output_dir/processed_output/deduped-statediff.sql
    ```

### CSV

* Create an env file with the required variables. Refer [.sample.env](./scripts/.sample.env).

* (Optional) Get row counts in the output:

    ```bash
    ./scripts/count-lines.sh <ENV_FILE_PATH>
    ```

* De-duplicate data:

    ```bash
    ./scripts/dedup.sh <ENV_FILE_PATH>
    ```

* Perform column checks:

    ```bash
    ./scripts/check-columns.sh <ENV_FILE_PATH>
    ```

    Check the output logs for any rows detected with unexpected number of columns.

    Example:

    ```bash
    # log
    eth.header_cids
    Start: Wednesday 21 September 2022 06:00:38 PM IST
    Time taken: 00:00:05
    End: Wednesday 21 September 2022 06:00:43 PM IST
    Total bad rows: 1 ./check-columns/eth.header_cids.txt

    # bad row output
    # line number, num. of columns, data
    23 17 22,xxxxxx,0x07f5ea5c94aa8dea60b28f6b6315d92f2b6d78ca4b74ea409adeb191b5a114f2,0x5918487321aa57dd0c50977856c6231e7c4ee79e95b694c7c8830227d77a1ecc,bagiacgzaa726uxeuvkg6uyfsr5vwgfozf4vw26gkjn2ouqe232yzdnnbctza,45,geth,0,0xad8fa8df61b98dbda7acd6ca76d5ce4cbba663d5f608cc940957adcdb94cee8d,0xc621412320a20b4aaff5363bdf063b9d13e394ef82e55689ab703aae5db08e26,0x71ec1c7d81269ce115be81c81f13e1cc2601c292a7f20440a77257ecfdc69940,0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347,\x2000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000000000,1658408419,/blocks/DMQAP5PKLSKKVDPKMCZI623DCXMS6K3NPDFEW5HKICNN5MMRWWQRJ4Q,1,0x0000000000000000000000000000000000000000
    ```

* Import data using `timescaledb-parallel-copy`:  
  (requires [`timescaledb-parallel-copy`](https://github.com/timescale/timescaledb-parallel-copy) installation; readily comes with TimescaleDB docker image)

    ```bash
    ./scripts/timescaledb-import.sh <ENV_FILE_PATH>
    ```

* NOTE: `COPY` command on CSVs inserts empty strings as `NULL` in the DB. Passing `FORCE_NOT_NULL <COLUMN_NAME>` forces it to insert empty strings instead. This is required to maintain compatibility of the imported statediff data with the data generated in `postgres` mode. Reference: https://www.postgresql.org/docs/14/sql-copy.html
