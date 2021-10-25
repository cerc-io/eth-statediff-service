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

const (
	statsSubsystem = "stats"
	subsystemHTTP  = "http"
	subsystemIPC   = "ipc"
)

var (
	metrics bool

	queuedRanges        prometheus.Gauge
	lastLoadedHeight    prometheus.Gauge
	lastProcessedHeight prometheus.Gauge

	tBlockLoad       prometheus.Histogram
	tBlockProcessing prometheus.Histogram
	tStateProcessing prometheus.Histogram
	tTxCommit        prometheus.Histogram

	httpCount    prometheus.Counter
	httpDuration prometheus.Histogram
	ipcCount     prometheus.Gauge
)

const (
	RANGES_QUEUED        = "ranges_queued"
	LOADED_HEIGHT        = "loaded_height"
	PROCESSED_HEIGHT     = "processed_height"
	T_BLOCK_LOAD         = "t_block_load"
	T_BLOCK_PROCESSING   = "t_block_processing"
	T_STATE_PROCESSING   = "t_state_processing"
	T_POSTGRES_TX_COMMIT = "t_postgres_tx_commit"
)

// Init module initialization
func Init() {
	metrics = true

	queuedRanges = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      RANGES_QUEUED,
		Help:      "Number of range requests currently queued",
	})
	lastLoadedHeight = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      LOADED_HEIGHT,
		Help:      "The last block that was loaded for processing",
	})
	lastProcessedHeight = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      PROCESSED_HEIGHT,
		Help:      "The last block that was processed",
	})

	tBlockLoad = promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: statsSubsystem,
		Name:      T_BLOCK_LOAD,
		Help:      "Block loading time",
	})
	tBlockProcessing = promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: statsSubsystem,
		Name:      T_BLOCK_PROCESSING,
		Help:      "Block (header, uncles, txs, rcts, tx trie, rct trie) processing time",
	})
	tStateProcessing = promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: statsSubsystem,
		Name:      T_STATE_PROCESSING,
		Help:      "State (state trie, storage tries, and code) processing time",
	})
	tTxCommit = promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: statsSubsystem,
		Name:      T_POSTGRES_TX_COMMIT,
		Help:      "Postgres tx commit time",
	})

	httpCount = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystemHTTP,
		Name:      "count",
		Help:      "http request count",
	})
	httpDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: subsystemHTTP,
		Name:      "duration",
		Help:      "http request duration",
	})
	ipcCount = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystemIPC,
		Name:      "count",
		Help:      "unix socket connection count",
	})
}

// RegisterDBCollector create metric collector for given connection
func RegisterDBCollector(name string, db *sqlx.DB) {
	if metrics {
		prometheus.Register(NewDBStatsCollector(name, db))
	}
}

// IncQueuedRanges increments the number of queued range requests
func IncQueuedRanges() {
	if metrics {
		queuedRanges.Inc()
	}
}

// DecQueuedRanges decrements the number of queued range requests
func DecQueuedRanges() {
	if metrics {
		queuedRanges.Dec()
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
	case T_BLOCK_LOAD:
		tBlockLoad.Observe(tAsF64)
	case T_BLOCK_PROCESSING:
		tBlockProcessing.Observe(tAsF64)
	case T_STATE_PROCESSING:
		tStateProcessing.Observe(tAsF64)
	case T_POSTGRES_TX_COMMIT:
		tTxCommit.Observe(tAsF64)
	}
}
