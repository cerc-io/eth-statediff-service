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
	"fmt"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/statediff/indexer/node"
	"github.com/ethereum/go-ethereum/statediff/indexer/postgres"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/vulcanize/eth-statediff-service/pkg/prom"
)

var (
	cfgFile        string
	subCommand     string
	logWithCommand log.Entry
)

var rootCmd = &cobra.Command{
	Use:              "eth-statediff-service",
	PersistentPreRun: initFuncs,
}

func Execute() {
	log.Info("----- Starting vDB -----")
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func initFuncs(cmd *cobra.Command, args []string) {
	logfile := viper.GetString("log.file")
	if logfile != "" {
		file, err := os.OpenFile(logfile,
			os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err == nil {
			log.Infof("Directing output to %s", logfile)
			log.SetOutput(file)
		} else {
			log.SetOutput(os.Stdout)
			log.Info("Failed to log to file, using default stdout")
		}
	} else {
		log.SetOutput(os.Stdout)
	}
	if err := logLevel(); err != nil {
		log.Fatal("Could not set log level: ", err)
	}

	if viper.GetBool("prom.metrics") {
		log.Info("initializing prometheus metrics")
		prom.Init()
	}

	if viper.GetBool("prom.http") {
		addr := fmt.Sprintf(
			"%s:%s",
			viper.GetString("prom.httpAddr"),
			viper.GetString("prom.httpPort"),
		)
		log.Info("starting prometheus server")
		prom.Listen(addr)
	}
}

func logLevel() error {
	viper.BindEnv("log.level", "LOGRUS_LEVEL")
	lvl, err := log.ParseLevel(viper.GetString("log.level"))
	if err != nil {
		return err
	}
	log.SetLevel(lvl)
	if lvl > log.InfoLevel {
		log.SetReportCaller(true)
	}
	log.Info("Log level set to ", lvl.String())
	return nil
}

func init() {
	cobra.OnInitialize(initConfig)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	rootCmd.PersistentFlags().String("http-path", "", "vdb server http path")
	rootCmd.PersistentFlags().String("ipc-path", "", "vdb server ipc path")
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file location")
	rootCmd.PersistentFlags().String("log-file", "", "file path for logging")
	rootCmd.PersistentFlags().String("log-level", log.InfoLevel.String(),
		"log level (trace, debug, info, warn, error, fatal, panic")

	rootCmd.PersistentFlags().String("leveldb-path", "", "path to primary datastore")
	rootCmd.PersistentFlags().String("ancient-path", "", "path to ancient datastore")

	rootCmd.PersistentFlags().Bool("prerun", false, "turn on prerun of toml configured ranges")
	rootCmd.PersistentFlags().Int("service-workers", 0, "number of range requests to process concurrently")
	rootCmd.PersistentFlags().Int("trie-workers", 0, "number of workers to use for trie traversal and processing")
	rootCmd.PersistentFlags().Int("worker-queue-size", 0, "size of the range request queue for service workers")

	rootCmd.PersistentFlags().String("database-name", "vulcanize_public", "database name")
	rootCmd.PersistentFlags().Int("database-port", 5432, "database port")
	rootCmd.PersistentFlags().String("database-hostname", "localhost", "database hostname")
	rootCmd.PersistentFlags().String("database-user", "", "database user")
	rootCmd.PersistentFlags().String("database-password", "", "database password")

	rootCmd.PersistentFlags().String("eth-node-id", "", "eth node id")
	rootCmd.PersistentFlags().String("eth-client-name", "Geth", "eth client name")
	rootCmd.PersistentFlags().String("eth-genesis-block",
		"0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3", "eth genesis block hash")
	rootCmd.PersistentFlags().String("eth-network-id", "1", "eth network id")
	rootCmd.PersistentFlags().String("eth-chain-id", "1", "eth chain id")

	rootCmd.PersistentFlags().Int("cache-db", 1024, "megabytes of memory allocated to database cache")
	rootCmd.PersistentFlags().Int("cache-trie", 1024, "Megabytes of memory allocated to trie cache")

	rootCmd.PersistentFlags().Bool("prom-http", false, "enable prometheus http service")
	rootCmd.PersistentFlags().String("prom-http-addr", "127.0.0.1", "prometheus http host")
	rootCmd.PersistentFlags().String("prom-http-port", "8080", "prometheus http port")

	rootCmd.PersistentFlags().Bool("metrics", false, "enable metrics")

	viper.BindPFlag("server.httpPath", rootCmd.PersistentFlags().Lookup("http-path"))
	viper.BindPFlag("server.ipcPath", rootCmd.PersistentFlags().Lookup("ipc-path"))
	viper.BindPFlag("log.file", rootCmd.PersistentFlags().Lookup("log-file"))
	viper.BindPFlag("log.level", rootCmd.PersistentFlags().Lookup("log-level"))
	viper.BindPFlag("statediff.prerun", rootCmd.PersistentFlags().Lookup("prerun"))
	viper.BindPFlag("statediff.serviceWorkers", rootCmd.PersistentFlags().Lookup("service-workers"))
	viper.BindPFlag("statediff.trieWorkers", rootCmd.PersistentFlags().Lookup("trie-workers"))
	viper.BindPFlag("statediff.workerQueueSize", rootCmd.PersistentFlags().Lookup("worker-queue-size"))
	viper.BindPFlag("leveldb.path", rootCmd.PersistentFlags().Lookup("leveldb-path"))
	viper.BindPFlag("leveldb.ancient", rootCmd.PersistentFlags().Lookup("ancient-path"))
	viper.BindPFlag("database.name", rootCmd.PersistentFlags().Lookup("database-name"))
	viper.BindPFlag("database.port", rootCmd.PersistentFlags().Lookup("database-port"))
	viper.BindPFlag("database.hostname", rootCmd.PersistentFlags().Lookup("database-hostname"))
	viper.BindPFlag("database.user", rootCmd.PersistentFlags().Lookup("database-user"))
	viper.BindPFlag("database.password", rootCmd.PersistentFlags().Lookup("database-password"))
	viper.BindPFlag("ethereum.nodeID", rootCmd.PersistentFlags().Lookup("eth-node-id"))
	viper.BindPFlag("ethereum.clientName", rootCmd.PersistentFlags().Lookup("eth-client-name"))
	viper.BindPFlag("ethereum.genesisBlock", rootCmd.PersistentFlags().Lookup("eth-genesis-block"))
	viper.BindPFlag("ethereum.networkID", rootCmd.PersistentFlags().Lookup("eth-network-id"))
	viper.BindPFlag("ethereum.chainID", rootCmd.PersistentFlags().Lookup("eth-chain-id"))
	viper.BindPFlag("cache.database", rootCmd.PersistentFlags().Lookup("cache-db"))
	viper.BindPFlag("cache.trie", rootCmd.PersistentFlags().Lookup("cache-trie"))
	viper.BindPFlag("prom.http", rootCmd.PersistentFlags().Lookup("prom-http"))
	viper.BindPFlag("prom.httpAddr", rootCmd.PersistentFlags().Lookup("prom-http-addr"))
	viper.BindPFlag("prom.httpPort", rootCmd.PersistentFlags().Lookup("prom-http-port"))
	viper.BindPFlag("prom.metrics", rootCmd.PersistentFlags().Lookup("metrics"))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
		if err := viper.ReadInConfig(); err == nil {
			log.Printf("Using config file: %s", viper.ConfigFileUsed())
		} else {
			log.Fatal(fmt.Sprintf("Couldn't read config file: %s", err.Error()))
		}
	} else {
		log.Warn("No config file passed with --config flag")
	}
}

func GetEthNodeInfo() node.Info {
	return node.Info{
		ID:           viper.GetString("ethereum.nodeID"),
		ClientName:   viper.GetString("ethereum.clientName"),
		GenesisBlock: viper.GetString("ethereum.genesisBlock"),
		NetworkID:    viper.GetString("ethereum.networkID"),
		ChainID:      viper.GetUint64("ethereum.chainID"),
	}
}

func GetDBParams() postgres.ConnectionParams {
	return postgres.ConnectionParams{
		Name:     viper.GetString("database.name"),
		Hostname: viper.GetString("database.hostname"),
		Port:     viper.GetInt("database.port"),
		User:     viper.GetString("database.user"),
		Password: viper.GetString("database.password"),
	}
}

func GetDBConfig() postgres.ConnectionConfig {
	return postgres.ConnectionConfig{
		MaxIdle:     viper.GetInt("database.maxIdle"),
		MaxOpen:     viper.GetInt("database.maxOpen"),
		MaxLifetime: viper.GetInt("database.maxLifetime"),
	}
}
