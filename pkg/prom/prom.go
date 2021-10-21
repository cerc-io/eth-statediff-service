// VulcanizeDB
// Copyright © 2021 Vulcanize

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

package prom

import (
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const statsSubsystem = "stats"

var (
	metrics bool

	lastLoadedHeight    prometheus.Gauge
	lastProcessedHeight prometheus.Gauge

	tBlockLoad       prometheus.Histogram
	tBlockProcessing prometheus.Histogram
	tStateProcessing prometheus.Histogram
	tTxCommit        prometheus.Histogram
)

// Init module initialization
func Init() {
	metrics = true

	lastLoadedHeight = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "loaded_height",
		Help:      "The last block that was loaded for processing",
	})
	lastProcessedHeight = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "processed_height",
		Help:      "The last block that was processed",
	})

	tBlockLoad = promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: statsSubsystem,
		Name:      "t_block_load",
		Help:      "Block loading time",
	})
	tBlockProcessing = promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: statsSubsystem,
		Name:      "t_block_processing",
		Help:      "Block (header, uncles, txs, rcts, tx trie, rct trie) processing time",
	})
	tStateProcessing = promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: statsSubsystem,
		Name:      "t_state_processing",
		Help:      "State (state trie, storage tries, and code) processing time",
	})
	tTxCommit = promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: statsSubsystem,
		Name:      "t_postgres_tx_commit",
		Help:      "Postgres tx commit time",
	})
}

// RegisterDBCollector create metric colletor for given connection
func RegisterDBCollector(name string, db *sqlx.DB) {
	if metrics {
		prometheus.Register(NewDBStatsCollector(name, db))
	}
}

// SetLastLoadedHeight sets last loaded height
func SetLastLoadedHeight(height int64) {
	if metrics {
		lastLoadedHeight.Set(float64(height))
	}
}

// SetLastProcessedHeight sets last processed height
func SetLastProcessedHeight(height int64) {
	if metrics {
		lastProcessedHeight.Set(float64(height))
	}
}

// SetTimeMetric time metric observation
func SetTimeMetric(name string, t time.Duration) {
	if !metrics {
		return
	}
	tAsF64 := t.Seconds()
	switch name {
	case "t_block_load":
		tBlockLoad.Observe(tAsF64)
	case "t_block_processing":
		tBlockProcessing.Observe(tAsF64)
	case "t_state_processing":
		tStateProcessing.Observe(tAsF64)
	case "t_postgres_tx_commit":
		tTxCommit.Observe(tAsF64)
	}
}
