// Copyright © 2022 Vulcanize, Inc
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
)

// statsCmd represents the serve command
var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Report stats for cold levelDB",
	Long: `Usage

./eth-statediff-service stats --config={path to toml config file}`,
	Run: func(cmd *cobra.Command, args []string) {
		subCommand = cmd.CalledAs()
		logWithCommand = *logrus.WithField("SubCommand", subCommand)
		stats()
	},
}

func init() {
	rootCmd.AddCommand(statsCmd)
}

func stats() {
	logWithCommand.Info("Running eth-statediff-service stats command")

	reader, _, _ := instantiateLevelDBReader()
	reportLatestBlock(reader)
}
