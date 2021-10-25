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
	"os"
	"os/signal"
	"sync"

	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	sd "github.com/vulcanize/eth-statediff-service/pkg"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Stand up a standalone statediffing RPC service on top of leveldb",
	Long: `Usage

./eth-statediff-service serve --config={path to toml config file}`,
	Run: func(cmd *cobra.Command, args []string) {
		subCommand = cmd.CalledAs()
		logWithCommand = *logrus.WithField("SubCommand", subCommand)
		serve()
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}

func serve() {
	logWithCommand.Info("Running eth-statediff-service serve command")

	statediffService, err := createStateDiffService()
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

func startServers(serv sd.StateDiffService) error {
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
