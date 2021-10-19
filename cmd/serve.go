// Copyright © 2019 Vulcanize, Inc
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"sync"

	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	ind "github.com/ethereum/go-ethereum/statediff/indexer"
	"github.com/ethereum/go-ethereum/statediff/indexer/postgres"

	sd "github.com/vulcanize/eth-statediff-service/pkg"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Standup a standalone statediffing RPC service on top of leveldb",
	Long: `Usage

./eth-statediff-service serve --config={path to toml config file}`,
	Run: func(cmd *cobra.Command, args []string) {
		subCommand = cmd.CalledAs()
		logWithCommand = *logrus.WithField("SubCommand", subCommand)
		serve()
	},
}

func serve() {
	logWithCommand.Info("starting statediff RPC service")

	// load params
	path := viper.GetString("leveldb.path")
	ancientPath := viper.GetString("leveldb.ancient")
	if path == "" || ancientPath == "" {
		logWithCommand.Fatal("require a valid eth leveldb primary datastore path and ancient datastore path")
	}

	nodeInfo := GetEthNodeInfo()
	config, err := chainConfig(nodeInfo.ChainID)
	if err != nil {
		logWithCommand.Fatal(err)
	}

	// create leveldb reader
	logWithCommand.Info("Creating leveldb reader")
	conf := sd.ReaderConfig{
		TrieConfig: &trie.Config{
			Cache:     viper.GetInt("cache.trie"),
			Journal:   "",
			Preimages: false,
		},
		ChainConfig: config,
		Path:        path,
		AncientPath: ancientPath,
		DBCacheSize: viper.GetInt("cache.database"),
	}
	lvlDBReader, err := sd.NewLvlDBReader(conf)
	if err != nil {
		logWithCommand.Fatal(err)
	}

	// create statediff service
	logWithCommand.Info("Creating statediff service")
	db, err := postgres.NewDB(postgres.DbConnectionString(GetDBParams()), GetDBConfig(), nodeInfo)
	if err != nil {
		logWithCommand.Fatal(err)
	}

	indexer, err := ind.NewStateDiffIndexer(config, db)
	if err != nil {
		logWithCommand.Fatal(err)
	}

	statediffService, err := sd.NewStateDiffService(lvlDBReader, indexer, viper.GetUint("statediff.workers"))
	if err != nil {
		logWithCommand.Fatal(err)
	}

	// start service and servers
	logWithCommand.Info("Starting statediff service")
	wg := new(sync.WaitGroup)
	go statediffService.Loop(wg)
	logWithCommand.Info("Starting RPC servers")
	if err := startServers(statediffService); err != nil {
		logWithCommand.Fatal(err)
	}
	logWithCommand.Info("RPC servers successfully spun up; awaiting requests")

	// clean shutdown
	shutdown := make(chan os.Signal)
	signal.Notify(shutdown, os.Interrupt)
	<-shutdown
	logWithCommand.Info("Received interrupt signal, shutting down")
	statediffService.Stop()
	wg.Wait()
}

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.PersistentFlags().String("http-path", "", "vdb server http path")
	serveCmd.PersistentFlags().String("ipc-path", "", "vdb server ipc path")

	viper.BindPFlag("server.httpPath", serveCmd.PersistentFlags().Lookup("http-path"))
	viper.BindPFlag("server.ipcPath", serveCmd.PersistentFlags().Lookup("ipc-path"))
}

func startServers(serv sd.IService) error {
	viper.BindEnv("server.ipcPath", "SERVER_IPC_PATH")
	viper.BindEnv("server.httpPath", "SERVER_HTTP_PATH")
	ipcPath := viper.GetString("server.ipcPath")
	httpPath := viper.GetString("server.httpPath")
	if ipcPath == "" && httpPath == "" {
		logWithCommand.Fatal("need an ipc path and/or an http path")
	}
	if ipcPath != "" {
		logWithCommand.Info("starting up IPC server")
		_, _, err := rpc.StartIPCEndpoint(ipcPath, serv.APIs())
		if err != nil {
			return err
		}
	}
	if httpPath != "" {
		logWithCommand.Info("starting up HTTP server")
		handler := rpc.NewServer()
		if _, _, err := node.StartHTTPEndpoint(httpPath, rpc.HTTPTimeouts{}, handler); err != nil {
			return err
		}
	}
	return nil
}

func chainConfig(chainID uint64) (*params.ChainConfig, error) {
	switch chainID {
	case 1:
		return params.MainnetChainConfig, nil
	case 3:
		return params.RopstenChainConfig, nil // Ropsten
	case 4:
		return params.RinkebyChainConfig, nil
	case 5:
		return params.GoerliChainConfig, nil
	default:
		return nil, fmt.Errorf("chain config for chainid %d not available", chainID)
	}
}
