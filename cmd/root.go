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
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/statediff/indexer/database/dump"
	"github.com/ethereum/go-ethereum/statediff/indexer/database/file"
	"github.com/ethereum/go-ethereum/statediff/indexer/database/sql/postgres"
	"github.com/ethereum/go-ethereum/statediff/indexer/interfaces"
	"github.com/ethereum/go-ethereum/statediff/indexer/node"
	"github.com/ethereum/go-ethereum/statediff/indexer/shared"
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

	rootCmd.PersistentFlags().String("leveldb-mode", "local", "LevelDB access mode (local, remote)")
	rootCmd.PersistentFlags().String("leveldb-path", "", "path to primary datastore")
	rootCmd.PersistentFlags().String("ancient-path", "", "path to ancient datastore")
	rootCmd.PersistentFlags().String("leveldb-url", "", "url to primary leveldb-ethdb-rpc server")

	rootCmd.PersistentFlags().Bool("prerun", false, "turn on prerun of toml configured ranges")
	rootCmd.PersistentFlags().Int("service-workers", 0, "number of range requests to process concurrently")
	rootCmd.PersistentFlags().Int("trie-workers", 0, "number of workers to use for trie traversal and processing")
	rootCmd.PersistentFlags().Int("worker-queue-size", 0, "size of the range request queue for service workers")

	rootCmd.PersistentFlags().String("database-name", "vulcanize_public", "database name")
	rootCmd.PersistentFlags().Int("database-port", 5432, "database port")
	rootCmd.PersistentFlags().String("database-hostname", "localhost", "database hostname")
	rootCmd.PersistentFlags().String("database-user", "", "database user")
	rootCmd.PersistentFlags().String("database-password", "", "database password")
	rootCmd.PersistentFlags().Int("database-max-idle", 0, "database max number of idle connections")
	rootCmd.PersistentFlags().Int("database-max-open", 0, "database max number of open connections")
	rootCmd.PersistentFlags().Int("database-min-open", 0, "database min number of open connections")
	rootCmd.PersistentFlags().Duration("database-max-conn-lifetime", 0, "database max connection lifetime")
	rootCmd.PersistentFlags().Duration("database-conn-timeout", 0, "database connection timeout")
	rootCmd.PersistentFlags().Duration("database-max-idle-time", 0, "database max connection idle time")
	rootCmd.PersistentFlags().String("database-type", "postgres", "database type (currently supported: postgres, dump)")
	rootCmd.PersistentFlags().String("database-driver", "sqlx", "database driver type (currently supported: sqlx, pgx)")
	rootCmd.PersistentFlags().String("database-dump-dst", "stdout", "dump destination (for database-type=dump; options: stdout, stderr, discard)")
	rootCmd.PersistentFlags().String("database-file-path", "", "full file path (for database-type=file)")

	rootCmd.PersistentFlags().String("eth-node-id", "", "eth node id")
	rootCmd.PersistentFlags().String("eth-client-name", "eth-statediff-service", "eth client name")
	rootCmd.PersistentFlags().String("eth-genesis-block",
		"0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3", "eth genesis block hash")
	rootCmd.PersistentFlags().String("eth-network-id", "1", "eth network id")
	rootCmd.PersistentFlags().String("eth-chain-id", "1", "eth chain id")

	rootCmd.PersistentFlags().Int("cache-db", 1024, "megabytes of memory allocated to database cache")
	rootCmd.PersistentFlags().Int("cache-trie", 1024, "Megabytes of memory allocated to trie cache")

	rootCmd.PersistentFlags().Bool("prom-http", false, "enable prometheus http service")
	rootCmd.PersistentFlags().String("prom-http-addr", "127.0.0.1", "prometheus http host")
	rootCmd.PersistentFlags().String("prom-http-port", "8080", "prometheus http port")
	rootCmd.PersistentFlags().Bool("prom-db-stats", false, "enables prometheus db stats")
	rootCmd.PersistentFlags().Bool("prom-metrics", false, "enable prometheus metrics")

	rootCmd.PersistentFlags().Bool("prerun-only", false, "only process pre-configured ranges; exit afterwards")
	rootCmd.PersistentFlags().Int("prerun-start", 0, "start height for a prerun range")
	rootCmd.PersistentFlags().Int("prerun-stop", 0, "stop height for a prerun range")
	rootCmd.PersistentFlags().Bool("prerun-intermediate-state-nodes", true, "include intermediate state nodes in state diff")
	rootCmd.PersistentFlags().Bool("prerun-intermediate-storage-nodes", true, "include intermediate storage nodes in state diff")
	rootCmd.PersistentFlags().Bool("prerun-include-block", true, "include block data in the statediff payload")
	rootCmd.PersistentFlags().Bool("prerun-include-receipts", true, "include receipts in the statediff payload")
	rootCmd.PersistentFlags().Bool("prerun-include-td", true, "include td in the statediff payload")
	rootCmd.PersistentFlags().Bool("prerun-include-code", true, "include code and codehash mappings in statediff payload")

	viper.BindPFlag("server.httpPath", rootCmd.PersistentFlags().Lookup("http-path"))
	viper.BindPFlag("server.ipcPath", rootCmd.PersistentFlags().Lookup("ipc-path"))

	viper.BindPFlag("log.file", rootCmd.PersistentFlags().Lookup("log-file"))
	viper.BindPFlag("log.level", rootCmd.PersistentFlags().Lookup("log-level"))

	viper.BindPFlag("statediff.prerun", rootCmd.PersistentFlags().Lookup("prerun"))
	viper.BindPFlag("statediff.serviceWorkers", rootCmd.PersistentFlags().Lookup("service-workers"))
	viper.BindPFlag("statediff.trieWorkers", rootCmd.PersistentFlags().Lookup("trie-workers"))
	viper.BindPFlag("statediff.workerQueueSize", rootCmd.PersistentFlags().Lookup("worker-queue-size"))

	viper.BindPFlag("leveldb.mode", rootCmd.PersistentFlags().Lookup("leveldb-mode"))
	viper.BindPFlag("leveldb.path", rootCmd.PersistentFlags().Lookup("leveldb-path"))
	viper.BindPFlag("leveldb.ancient", rootCmd.PersistentFlags().Lookup("ancient-path"))
	viper.BindPFlag("leveldb.url", rootCmd.PersistentFlags().Lookup("leveldb-url"))

	viper.BindPFlag("database.name", rootCmd.PersistentFlags().Lookup("database-name"))
	viper.BindPFlag("database.port", rootCmd.PersistentFlags().Lookup("database-port"))
	viper.BindPFlag("database.hostname", rootCmd.PersistentFlags().Lookup("database-hostname"))
	viper.BindPFlag("database.user", rootCmd.PersistentFlags().Lookup("database-user"))
	viper.BindPFlag("database.password", rootCmd.PersistentFlags().Lookup("database-password"))
	viper.BindPFlag("database.maxIdle", rootCmd.PersistentFlags().Lookup("database-max-idle"))
	viper.BindPFlag("database.maxOpen", rootCmd.PersistentFlags().Lookup("database-max-open"))
	viper.BindPFlag("database.minOpen", rootCmd.PersistentFlags().Lookup("database-min-open"))
	viper.BindPFlag("database.maxConnLifetime", rootCmd.PersistentFlags().Lookup("database-max-conn-lifetime"))
	viper.BindPFlag("database.connTimeout", rootCmd.PersistentFlags().Lookup("database-conn-timeout"))
	viper.BindPFlag("database.maxIdleTime", rootCmd.PersistentFlags().Lookup("database-max-idle-time"))
	viper.BindPFlag("database.type", rootCmd.PersistentFlags().Lookup("database-type"))
	viper.BindPFlag("database.driver", rootCmd.PersistentFlags().Lookup("database-driver"))
	viper.BindPFlag("database.dumpDestination", rootCmd.PersistentFlags().Lookup("database-dump-dst"))
	viper.BindPFlag("database.filePath", rootCmd.PersistentFlags().Lookup("database-file-path"))

	viper.BindPFlag("ethereum.nodeID", rootCmd.PersistentFlags().Lookup("eth-node-id"))
	viper.BindPFlag("ethereum.clientName", rootCmd.PersistentFlags().Lookup("eth-client-name"))
	viper.BindPFlag("ethereum.genesisBlock", rootCmd.PersistentFlags().Lookup("eth-genesis-block"))
	viper.BindPFlag("ethereum.networkID", rootCmd.PersistentFlags().Lookup("eth-network-id"))
	viper.BindPFlag("ethereum.chainID", rootCmd.PersistentFlags().Lookup("eth-chain-id"))
	viper.BindPFlag("ethereum.chainConfig", rootCmd.PersistentFlags().Lookup("eth-chain-config"))

	viper.BindPFlag("cache.database", rootCmd.PersistentFlags().Lookup("cache-db"))
	viper.BindPFlag("cache.trie", rootCmd.PersistentFlags().Lookup("cache-trie"))

	viper.BindPFlag("prom.http", rootCmd.PersistentFlags().Lookup("prom-http"))
	viper.BindPFlag("prom.httpAddr", rootCmd.PersistentFlags().Lookup("prom-http-addr"))
	viper.BindPFlag("prom.httpPort", rootCmd.PersistentFlags().Lookup("prom-http-port"))
	viper.BindPFlag("prom.dbStats", rootCmd.PersistentFlags().Lookup("prom-db-stats"))
	viper.BindPFlag("prom.metrics", rootCmd.PersistentFlags().Lookup("prom-metrics"))

	viper.BindPFlag("prerun.only", rootCmd.PersistentFlags().Lookup("prerun-only"))
	viper.BindPFlag("prerun.start", rootCmd.PersistentFlags().Lookup("prerun-start"))
	viper.BindPFlag("prerun.stop", rootCmd.PersistentFlags().Lookup("prerun-stop"))
	viper.BindPFlag("prerun.params.intermediateStateNodes", rootCmd.PersistentFlags().Lookup("prerun-intermediate-state-nodes"))
	viper.BindPFlag("prerun.params.intermediateStorageNodes", rootCmd.PersistentFlags().Lookup("prerun-intermediate-storage-nodes"))
	viper.BindPFlag("prerun.params.includeBlock", rootCmd.PersistentFlags().Lookup("prerun-include-block"))
	viper.BindPFlag("prerun.params.includeReceipts", rootCmd.PersistentFlags().Lookup("prerun-include-receipts"))
	viper.BindPFlag("prerun.params.includeTD", rootCmd.PersistentFlags().Lookup("prerun-include-td"))
	viper.BindPFlag("prerun.params.includeCode", rootCmd.PersistentFlags().Lookup("prerun-include-code"))

	rand.Seed(time.Now().UnixNano())
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

func getEthNodeInfo() node.Info {
	var nodeID, genesisBlock, networkID, clientName string
	var chainID uint64
	if !viper.IsSet("ethereum.nodeID") {
		nodeID = randSeq(12)
	} else {
		nodeID = viper.GetString("ethereum.nodeID")
	}
	genesisBlock = viper.GetString("ethereum.genesisBlock")
	if !viper.IsSet("ethereum.genesisBlock") {
		genesisBlock = "0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3"
	} else {
		genesisBlock = viper.GetString("ethereum.genesisBlock")
	}
	if !viper.IsSet("ethereum.chainID") {
		chainID = 1
	} else {
		chainID = viper.GetUint64("ethereum.chainID")
	}
	if !viper.IsSet("ethereum.networkID") {
		networkID = "1"
	} else {
		networkID = viper.GetString("ethereum.networkID")
	}
	if !viper.IsSet("ethereum.clientName") {
		clientName = "eth-statediff-service"
	} else {
		clientName = viper.GetString("ethereum.clientName")
	}
	return node.Info{
		ID:           nodeID,
		ClientName:   clientName,
		GenesisBlock: genesisBlock,
		NetworkID:    networkID,
		ChainID:      chainID,
	}
}

var characters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = characters[rand.Intn(len(characters))]
	}
	return string(b)
}

// getConfig constructs and returns the appropriate config from viper params
func getConfig(nodeInfo node.Info) (interfaces.Config, error) {
	dbTypeStr := viper.GetString("database.type")
	dbType, err := shared.ResolveDBType(dbTypeStr)
	if err != nil {
		return nil, err
	}
	logWithCommand.Infof("configuring service for database type: %s", dbType)
	var indexerConfig interfaces.Config
	switch dbType {
	case shared.FILE:
		logWithCommand.Info("starting in sql file writing mode")
		filePathStr := viper.GetString("database.filePath")
		if filePathStr == "" {
			logWithCommand.Fatal("when operating in sql file writing mode a file path must be provided")
		}
		indexerConfig = file.Config{FilePath: filePathStr}
	case shared.DUMP:
		logWithCommand.Info("starting in data dump mode")
		dumpDstStr := viper.GetString("database.dumpDestination")
		dumpDst, err := dump.ResolveDumpType(dumpDstStr)
		if err != nil {
			return nil, err
		}
		switch dumpDst {
		case dump.STDERR:
			indexerConfig = dump.Config{Dump: os.Stdout}
		case dump.STDOUT:
			indexerConfig = dump.Config{Dump: os.Stderr}
		case dump.DISCARD:
			indexerConfig = dump.Config{Dump: dump.NewDiscardWriterCloser()}
		default:
			return nil, fmt.Errorf("unrecognized dump destination: %s", dumpDst)
		}
	case shared.POSTGRES:
		logWithCommand.Info("starting in postgres mode")
		driverTypeStr := viper.GetString("database.driver")
		driverType, err := postgres.ResolveDriverType(driverTypeStr)
		if err != nil {
			utils.Fatalf("%v", err)
		}
		pgConfig := postgres.Config{
			Hostname:     viper.GetString("database.hostname"),
			Port:         viper.GetInt("database.port"),
			DatabaseName: viper.GetString("database.name"),
			Username:     viper.GetString("database.user"),
			Password:     viper.GetString("database.password"),
			ID:           nodeInfo.ID,
			ClientName:   nodeInfo.ClientName,
			Driver:       driverType,
		}
		if viper.IsSet("database.maxIdle") {
			pgConfig.MaxIdle = viper.GetInt("database.maxIdle")
		}
		if viper.IsSet("database.maxOpen") {
			pgConfig.MaxConns = viper.GetInt("database.maxOpen")
		}
		if viper.IsSet("database.minOpen") {
			pgConfig.MinConns = viper.GetInt("database.minOpen")
		}
		if viper.IsSet("database.maxConnLifetime") {
			pgConfig.MaxConnLifetime = viper.GetDuration("database.maxConnLifetime")
		}
		if viper.IsSet("database.connTimeout") {
			pgConfig.ConnTimeout = viper.GetDuration("database.connTimeout")
		}
		if viper.IsSet("database.maxIdleTime") {
			pgConfig.MaxConnIdleTime = viper.GetDuration("database.maxIdleTime")
		}
		indexerConfig = pgConfig
	default:
		return nil, fmt.Errorf("unrecognized db type: %s", dbType)
	}
	return indexerConfig, nil
}
