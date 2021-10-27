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

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// clientCmd represents the serve command
var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "Client for queuing range requests",
	Long: `Usage

./eth-statediff-service client --config={path to toml config file}`,
	Run: func(cmd *cobra.Command, args []string) {
		subCommand = cmd.CalledAs()
		logWithCommand = *logrus.WithField("SubCommand", subCommand)
		client()
	},
}

func init() {
	rootCmd.AddCommand(clientCmd)
}

func client() {
	logWithCommand.Info("Running eth-statediff-service client command")

	statediffService, err := createStateDiffService()
	if err != nil {
		logWithCommand.Fatal(err)
	}

	// start service and clientrs
	logWithCommand.Info("Starting statediff service")
	wg := new(sync.WaitGroup)
	if err := statediffService.Loop(wg); err != nil {
		logWithCommand.Fatalf("unable to start statediff service: %v", err)
	}
	logWithCommand.Info("Starting RPC clientrs")
	if err := startServers(statediffService); err != nil {
		logWithCommand.Fatal(err)
	}
	logWithCommand.Info("RPC clientrs successfully spun up; awaiting requests")

	// clean shutdown
	shutdown := make(chan os.Signal)
	signal.Notify(shutdown, os.Interrupt)
	<-shutdown
	logWithCommand.Info("Received interrupt signal, shutting down")
	statediffService.Stop()
	wg.Wait()
}
