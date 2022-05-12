// VulcanizeDB
// Copyright © 2019 Vulcanize

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.

// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package cmd

import (
	"github.com/spf13/viper"
)

const (
	ETH_NODE_ID       = "ETH_NODE_ID"
	ETH_CLIENT_NAME   = "ETH_CLIENT_NAME"
	ETH_GENESIS_BLOCK = "ETH_GENESIS_BLOCK"
	ETH_NETWORK_ID    = "ETH_NETWORK_ID"
	ETH_CHAIN_ID      = "ETH_CHAIN_ID"
	ETH_CHAIN_CONFIG  = "ETH_CHAIN_CONFIG"

	DB_CACHE_SIZE_MB   = "DB_CACHE_SIZE_MB"
	TRIE_CACHE_SIZE_MB = "TRIE_CACHE_SIZE_MB"
	LVLDB_PATH         = "LVLDB_PATH"
	LVLDB_ANCIENT      = "LVLDB_ANCIENT"

	STATEDIFF_PRERUN            = "STATEDIFF_PRERUN"
	STATEDIFF_TRIE_WORKERS      = "STATEDIFF_TRIE_WORKERS"
	STATEDIFF_SERVICE_WORKERS   = "STATEDIFF_SERVICE_WORKERS"
	STATEDIFF_WORKER_QUEUE_SIZE = "STATEDIFF_WORKER_QUEUE_SIZE"

	SERVICE_IPC_PATH  = "SERVICE_IPC_PATH"
	SERVICE_HTTP_PATH = "SERVICE_HTTP_PATH"

	PROM_METRICS   = "PROM_METRICS"
	PROM_HTTP      = "PROM_HTTP"
	PROM_HTTP_ADDR = "PROM_HTTP_ADDR"
	PROM_HTTP_PORT = "PROM_HTTP_PORT"
	PROM_DB_STATS  = "PROM_DB_STATS"

	PRERUN_ONLY                       = "PRERUN_ONLY"
	PRERUN_RANGE_START                = "PRERUN_RANGE_START"
	PRERUN_RANGE_STOP                 = "PRERUN_RANGE_STOP"
	PRERUN_INTERMEDIATE_STATE_NODES   = "PRERUN_INTERMEDIATE_STATE_NODES"
	PRERUN_INTERMEDIATE_STORAGE_NODES = "PRERUN_INTERMEDIATE_STORAGE_NODES"
	PRERUN_INCLUDE_BLOCK              = "PRERUN_INCLUDE_BLOCK"
	PRERUN_INCLUDE_RECEIPTS           = "PRERUN_INCLUDE_RECEIPTS"
	PRERUN_INCLUDE_TD                 = "PRERUN_INCLUDE_TD"
	PRERUN_INCLUDE_CODE               = "PRERUN_INCLUDE_CODE"

	LOG_LEVEL     = "LOG_LEVEL"
	LOG_FILE_PATH = "LOG_FILE_PATH"

	DATABASE_NAME     = "DATABASE_NAME"
	DATABASE_HOSTNAME = "DATABASE_HOSTNAME"
	DATABASE_PORT     = "DATABASE_PORT"
	DATABASE_USER     = "DATABASE_USER"
	DATABASE_PASSWORD = "DATABASE_PASSWORD"

	DATABASE_TYPE        = "DATABASE_TYPE"
	DATABASE_DRIVER_TYPE = "DATABASE_DRIVER_TYPE"
	DATABASE_DUMP_DST    = "DATABASE_DUMP_DST"
	DATABASE_FILE_PATH   = "DATABASE_FILE_PATH"

	DATABASE_MAX_IDLE_CONNECTIONS = "DATABASE_MAX_IDLE_CONNECTIONS"
	DATABASE_MAX_OPEN_CONNECTIONS = "DATABASE_MAX_OPEN_CONNECTIONS"
	DATABASE_MIN_OPEN_CONNS       = "DATABASE_MIN_OPEN_CONNS"
	DATABASE_MAX_CONN_LIFETIME    = "DATABASE_MAX_CONN_LIFETIME"
	DATABASE_CONN_TIMEOUT         = "DATABSE_CONN_TIMEOUT"
	DATABASE_MAX_CONN_IDLE_TIME   = "DATABASE_MAX_CONN_IDLE_TIME"
)

// Bind env vars for eth node and DB configuration
func init() {
	viper.BindEnv("server.ipcPath", SERVICE_IPC_PATH)
	viper.BindEnv("server.httpPath", SERVICE_HTTP_PATH)

	viper.BindEnv("ethereum.nodeID", ETH_NODE_ID)
	viper.BindEnv("ethereum.clientName", ETH_CLIENT_NAME)
	viper.BindEnv("ethereum.genesisBlock", ETH_GENESIS_BLOCK)
	viper.BindEnv("ethereum.networkID", ETH_NETWORK_ID)
	viper.BindEnv("ethereum.chainID", ETH_CHAIN_ID)
	viper.BindEnv("ethereum.chainConfig", ETH_CHAIN_CONFIG)

	viper.BindEnv("database.name", DATABASE_NAME)
	viper.BindEnv("database.hostname", DATABASE_HOSTNAME)
	viper.BindEnv("database.port", DATABASE_PORT)
	viper.BindEnv("database.user", DATABASE_USER)
	viper.BindEnv("database.password", DATABASE_PASSWORD)

	viper.BindEnv("database.maxIdle", DATABASE_MAX_IDLE_CONNECTIONS)
	viper.BindEnv("database.maxOpen", DATABASE_MAX_OPEN_CONNECTIONS)
	viper.BindEnv("database.minOpen", DATABASE_MIN_OPEN_CONNS)
	viper.BindEnv("database.maxConnLifetime", DATABASE_MAX_CONN_LIFETIME)
	viper.BindEnv("database.connTimeout", DATABASE_CONN_TIMEOUT)
	viper.BindEnv("database.maxIdleTime", DATABASE_MAX_CONN_IDLE_TIME)

	viper.BindEnv("database.type", DATABASE_TYPE)
	viper.BindEnv("database.driver", DATABASE_DRIVER_TYPE)
	viper.BindEnv("database.dumpDestination", DATABASE_DUMP_DST)
	viper.BindEnv("database.filePath", DATABASE_FILE_PATH)

	viper.BindEnv("cache.database", DB_CACHE_SIZE_MB)
	viper.BindEnv("cache.trie", TRIE_CACHE_SIZE_MB)

	viper.BindEnv("leveldb.path", LVLDB_PATH)
	viper.BindEnv("leveldb.ancient", LVLDB_ANCIENT)

	viper.BindEnv("prom.metrics", PROM_METRICS)
	viper.BindEnv("prom.http", PROM_HTTP)
	viper.BindEnv("prom.httpAddr", PROM_HTTP_ADDR)
	viper.BindEnv("prom.httpPort", PROM_HTTP_PORT)
	viper.BindEnv("prom.dbStats", PROM_DB_STATS)

	viper.BindEnv("statediff.serviceWorkers", STATEDIFF_SERVICE_WORKERS)
	viper.BindEnv("statediff.trieWorkers", STATEDIFF_TRIE_WORKERS)
	viper.BindEnv("statediff.workerQueueSize", STATEDIFF_WORKER_QUEUE_SIZE)

	viper.BindEnv("statediff.prerun", STATEDIFF_PRERUN)
	viper.BindEnv("prerun.only", PRERUN_ONLY)
	viper.BindEnv("prerun.start", PRERUN_RANGE_START)
	viper.BindEnv("prerun.stop", PRERUN_RANGE_STOP)
	viper.BindEnv("prerun.params.intermediateStateNodes", PRERUN_INTERMEDIATE_STATE_NODES)
	viper.BindEnv("prerun.params.intermediateStorageNodes", PRERUN_INTERMEDIATE_STORAGE_NODES)
	viper.BindEnv("prerun.params.includeBlock", PRERUN_INCLUDE_BLOCK)
	viper.BindEnv("prerun.params.includeReceipts", PRERUN_INCLUDE_RECEIPTS)
	viper.BindEnv("prerun.params.includeTD", PRERUN_INCLUDE_TD)
	viper.BindEnv("prerun.params.includeCode", PRERUN_INCLUDE_CODE)

	viper.BindEnv("log.level", LOG_LEVEL)
	viper.BindEnv("log.file", LOG_FILE_PATH)
}
