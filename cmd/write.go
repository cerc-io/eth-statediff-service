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
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	gethsd "github.com/ethereum/go-ethereum/statediff"
	ind "github.com/ethereum/go-ethereum/statediff/indexer"
	"github.com/ethereum/go-ethereum/statediff/indexer/postgres"
	sd "github.com/vulcanize/eth-statediff-service/pkg"
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

func init() {
	rootCmd.AddCommand(writeCmd)
}

func write() {
	logWithCommand.Info("Starting statediff writer")

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
	lvlDBReader, err := sd.NewLvlDBReader(path, ancientPath, config)
	if err != nil {
		logWithCommand.Fatal(err)
	}

	// create statediff service
	logWithCommand.Info("Creating statediff service")
	db, err := postgres.NewDB(postgres.DbConnectionString(GetDBParams()), GetDBConfig(), nodeInfo)
	if err != nil {
		logWithCommand.Fatal(err)
	}
	indexer := ind.NewStateDiffIndexer(config, db)
	statediffService, err := sd.NewStateDiffService(lvlDBReader, indexer, viper.GetUint("statediff.workers"))
	if err != nil {
		logWithCommand.Fatal(err)
	}

	// Read all defined block ranges, write statediffs to database
	var blockRanges [][2]uint64
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

	for _, rng := range blockRanges {
		if rng[1] < rng[0] {
			logWithCommand.Fatal("range ending block number needs to be greater than starting block number")
		}
		logrus.Infof("Writing statediffs from block %d to %d", rng[0], rng[1])
		for height := rng[0]; height <= rng[1]; height++ {
			statediffService.WriteStateDiffAt(height, diffParams)
		}
	}
}
