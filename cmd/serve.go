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

	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethereum/go-ethereum/statediff"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

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
	chainID := viper.GetUint64("eth.chainID")
	config, err := chainConfig(chainID)
	if err != nil {
		logWithCommand.Fatal(err)
	}

	// create leveldb reader
	logWithCommand.Info("creating leveldb reader")
	lvlDBReader, err := sd.NewLvlDBReader(path, ancientPath, config)
	if err != nil {
		logWithCommand.Fatal(err)
	}

	// create statediff service
	logWithCommand.Info("creating statediff service")
	statediffService, err := sd.NewStateDiffService(lvlDBReader)
	if err != nil {
		logWithCommand.Fatal(err)
	}

	// start service and servers
	logWithCommand.Info("starting statediff service")
	wg := new(sync.WaitGroup)
	go statediffService.Loop(wg)
	logWithCommand.Info("starting rpc servers")
	if err := startServers(statediffService); err != nil {
		logWithCommand.Fatal(err)
	}
	logWithCommand.Info("rpc servers successfully spun up; awaiting requests")

	// clean shutdown
	shutdown := make(chan os.Signal)
	signal.Notify(shutdown, os.Interrupt)
	<-shutdown
	logWithCommand.Info("received interrupt signal, shutting down")
	statediffService.Stop()
	wg.Wait()
}

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.PersistentFlags().String("leveldb-path", "", "path to primary datastore")
	serveCmd.PersistentFlags().String("ancient-path", "", "path to ancient datastore")
	serveCmd.PersistentFlags().Uint("chain-id", 1, "ethereum chain id (mainnet = 1)")
	serveCmd.PersistentFlags().String("http-path", "", "vdb server http path")
	serveCmd.PersistentFlags().String("ipc-path", "", "vdb server ipc path")

	viper.BindPFlag("leveldb.path", serveCmd.PersistentFlags().Lookup("leveldb-path"))
	viper.BindPFlag("leveldb.ancient", serveCmd.PersistentFlags().Lookup("ancient-path"))
	viper.BindPFlag("eth.chainID", serveCmd.PersistentFlags().Lookup("chain-id"))
	viper.BindPFlag("server.httpPath", serveCmd.PersistentFlags().Lookup("http-path"))
	viper.BindPFlag("server.ipcPath", serveCmd.PersistentFlags().Lookup("ipc-path"))
}

func startServers(serv sd.IService) error {
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
		if _, _, err := rpc.StartHTTPEndpoint(httpPath, serv.APIs(), []string{statediff.APIName}, nil, nil, rpc.HTTPTimeouts{}); err != nil {
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
		return params.TestnetChainConfig, nil // Ropsten
	case 4:
		return params.RinkebyChainConfig, nil
	case 5:
		return params.GoerliChainConfig, nil
	default:
		return nil, fmt.Errorf("chain config for chainid %d not available", chainID)
	}
}
