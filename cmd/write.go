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
	"net/http"
	"strconv"
	"sync"
	"time"

	gethsd "github.com/ethereum/go-ethereum/statediff"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var writeCmd = &cobra.Command{
	Use:   "write",
	Short: "Write statediffs directly to DB for preconfigured block ranges",
	Long: `Usage

./eth-statediff-service write --config={path to toml config file}`,
	Run: func(cmd *cobra.Command, args []string) {
		subCommand = cmd.CalledAs()
		logWithCommand = *logrus.WithField("SubCommand", subCommand)
		write()
	},
}

type blockRange [2]uint64

func init() {
	rootCmd.AddCommand(writeCmd)
	writeCmd.PersistentFlags().String("write-api", "", "starts a server which handles write request through endpoints")
	viper.BindPFlag("write.serve", writeCmd.PersistentFlags().Lookup("write-api"))
}

func write() {
	logWithCommand.Info("Running eth-statediff-service write command")
	viper.BindEnv("write.serve", WRITE_SERVER)
	addr := viper.GetString("write.serve")

	statediffService, err := createStateDiffService()
	if err != nil {
		logWithCommand.Fatal(err)
	}

	// Read all defined block ranges, write statediffs to database
	var blockRanges []blockRange
	diffParams := gethsd.Params{ // todo: configurable?
		IntermediateStateNodes:   true,
		IntermediateStorageNodes: true,
		IncludeBlock:             true,
		IncludeReceipts:          true,
		IncludeTD:                true,
		IncludeCode:              true,
	}
	viper.UnmarshalKey("write.ranges", &blockRanges)
	viper.UnmarshalKey("write.params", &diffParams)

	blockRangesCh := make(chan blockRange, 100)
	go func() {
		for _, r := range blockRanges {
			blockRangesCh <- r
		}
		if addr == "" {
			close(blockRangesCh)
			return
		}
		startServer(addr, blockRangesCh)
	}()

	processRanges(statediffService, diffParams, blockRangesCh)
}

func startServer(addr string, blockRangesCh chan<- blockRange) {
	handler := func(w http.ResponseWriter, req *http.Request) {
		start, err := strconv.Atoi(req.URL.Query().Get("start"))
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to parse start value: %v", err), http.StatusInternalServerError)
			return
		}

		end, err := strconv.Atoi(req.URL.Query().Get("end"))
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to parse end value: %v", err), http.StatusInternalServerError)
			return
		}

		select {
		case blockRangesCh <- blockRange{uint64(start), uint64(end)}:
		case <-time.After(time.Millisecond * 200):
			http.Error(w, "server is busy", http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "added block range to the queue\n")
	}

	http.HandleFunc("/writeDiff", handler)
	logrus.Fatal(http.ListenAndServe(addr, nil))
}

type diffService interface {
	WriteStateDiffAt(blockNumber uint64, params gethsd.Params) error
}

func processRanges(sds diffService, param gethsd.Params, blockRangesCh chan blockRange) {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for rng := range blockRangesCh {
			if rng[1] < rng[0] {
				logWithCommand.Fatal("range ending block number needs to be greater than starting block number")
			}
			logrus.Infof("Writing statediffs from block %d to %d", rng[0], rng[1])
			for height := rng[0]; height <= rng[1]; height++ {
				err := sds.WriteStateDiffAt(height, param)
				if err != nil {
					logrus.Errorf("failed to write state diff for range: %v %v", rng, err)
				}
			}
		}
	}()

	wg.Wait()
}
