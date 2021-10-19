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

	pg "github.com/ethereum/go-ethereum/statediff/indexer/postgres"
)

const (
	ETH_NODE_ID       = "ETH_NODE_ID"
	ETH_CLIENT_NAME   = "ETH_CLIENT_NAME"
	ETH_GENESIS_BLOCK = "ETH_GENESIS_BLOCK"
	ETH_NETWORK_ID    = "ETH_NETWORK_ID"
	ETH_CHAIN_ID      = "ETH_CHAIN_ID"

	DB_CACHE_SIZE_MB   = "DB_CACHE_SIZE_MB"
	TRIE_CACHE_SIZE_MB = "TRIE_CACHE_SIZE_MB"

	WRITE_SERVER = "WRITE_SERVER"

	STATEDIFF_WORKERS = "STATEDIFF_WORKERS"
)

// Bind env vars for eth node and DB configuration
func init() {
	viper.BindEnv("ethereum.nodeID", ETH_NODE_ID)
	viper.BindEnv("ethereum.clientName", ETH_CLIENT_NAME)
	viper.BindEnv("ethereum.genesisBlock", ETH_GENESIS_BLOCK)
	viper.BindEnv("ethereum.networkID", ETH_NETWORK_ID)
	viper.BindEnv("ethereum.chainID", ETH_CHAIN_ID)

	viper.BindEnv("database.name", pg.DATABASE_NAME)
	viper.BindEnv("database.hostname", pg.DATABASE_HOSTNAME)
	viper.BindEnv("database.port", pg.DATABASE_PORT)
	viper.BindEnv("database.user", pg.DATABASE_USER)
	viper.BindEnv("database.password", pg.DATABASE_PASSWORD)
	viper.BindEnv("database.maxIdle", pg.DATABASE_MAX_IDLE_CONNECTIONS)
	viper.BindEnv("database.maxOpen", pg.DATABASE_MAX_OPEN_CONNECTIONS)
	viper.BindEnv("database.maxLifetime", pg.DATABASE_MAX_CONN_LIFETIME)

	viper.BindEnv("cache.database", DB_CACHE_SIZE_MB)
	viper.BindEnv("cache.trie", TRIE_CACHE_SIZE_MB)

	viper.BindEnv("write.serve", WRITE_SERVER)

	viper.BindEnv("statediff.workers", STATEDIFF_WORKERS)
}
