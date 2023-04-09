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
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	sd "github.com/cerc-io/eth-statediff-service/pkg"
	srpc "github.com/cerc-io/eth-statediff-service/pkg/rpc"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Stand up a standalone statediffing RPC service on top of LevelDB",
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

func maxParallelism() int {
	maxProcs := runtime.GOMAXPROCS(0)
	numCPU := runtime.NumCPU()
	if maxProcs < numCPU {
		return maxProcs
	}
	return numCPU
}

func serve() {
	logWithCommand.Info("Running eth-statediff-service serve command")
	logWithCommand.Infof("Parallelism: %d", maxParallelism())

	reader, chainConf, nodeInfo := instantiateLevelDBReader()

	// report latest block info
	header, err := reader.GetLatestHeader()
	if err != nil {
		logWithCommand.Fatalf("Unable to determine latest header height and hash: %s", err.Error())
	}
	if header.Number == nil {
		logWithCommand.Fatal("Latest header found in levelDB has a nil block height")
	}
	logWithCommand.Infof("Latest block found in the levelDB\r\nheight: %s, hash: %s", header.Number.String(), header.Hash().Hex())

	statediffService, err := createStateDiffService(reader, chainConf, nodeInfo)
	if err != nil {
		logWithCommand.Fatal(err)
	}

	// Enable the pprof agent if configured
	if viper.GetBool("debug.pprof") {
		// See: https://www.farsightsecurity.com/blog/txt-record/go-remote-profiling-20161028/
		// For security reasons: do not use the default http multiplexor elsewhere in this process.
		go func() {
			logWithCommand.Info("Starting pprof listener on port 6060")
			logWithCommand.Fatal(http.ListenAndServe("localhost:6060", nil))
		}()
	}

	// short circuit if we only want to perform prerun
	if viper.GetBool("prerun.only") {
		parallel := viper.GetBool("prerun.parallel")
		if err := statediffService.Run(nil, parallel); err != nil {
			logWithCommand.Fatal("Unable to perform prerun: %v", err)
		}
		return
	}

	// start service and servers
	logWithCommand.Info("Starting statediff service")
	wg := new(sync.WaitGroup)
	if err := statediffService.Loop(wg); err != nil {
		logWithCommand.Fatalf("unable to start statediff service: %v", err)
	}
	logWithCommand.Info("Starting RPC servers")
	if err := startServers(statediffService); err != nil {
		logWithCommand.Fatal(err)
	}
	logWithCommand.Info("RPC servers successfully spun up; awaiting requests")

	// clean shutdown
	shutdown := make(chan os.Signal)
	signal.Notify(shutdown, syscall.SIGINT)
	<-shutdown
	logWithCommand.Info("Received interrupt signal, shutting down")
	statediffService.Stop()
	wg.Wait()
}

func startServers(serv sd.StateDiffService) error {
	ipcPath := viper.GetString("server.ipcPath")
	httpPath := viper.GetString("server.httpPath")
	if ipcPath == "" && httpPath == "" {
		logWithCommand.Fatal("Need an ipc path and/or an http path")
	}
	if ipcPath != "" {
		logWithCommand.Info("Starting up IPC server")
		_, _, err := srpc.StartIPCEndpoint(ipcPath, serv.APIs())
		if err != nil {
			return err
		}
	}
	if httpPath != "" {
		logWithCommand.Info("Starting up HTTP server")
		_, err := srpc.StartHTTPEndpoint(httpPath, serv.APIs(), []string{"statediff"}, nil, []string{"*"}, rpc.HTTPTimeouts{})
		if err != nil {
			return err
		}
	} else {
		logWithCommand.Info("HTTP server is disabled")
	}

	return nil
}
