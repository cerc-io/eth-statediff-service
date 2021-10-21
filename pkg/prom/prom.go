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

	receipts     prometheus.Counter
	transactions prometheus.Counter
	blocks       prometheus.Counter
	logs		prometheus.Counter
	accessListEntries prometheus.Counter

	lenPayloadChan prometheus.Gauge

	tPayloadDecode             prometheus.Histogram
	tFreePostgres              prometheus.Histogram
	tPostgresCommit            prometheus.Histogram
	tHeaderProcessing          prometheus.Histogram
	tUncleProcessing           prometheus.Histogram
	tTxAndRecProcessing        prometheus.Histogram
	tStateAndStoreProcessing   prometheus.Histogram
	tCodeAndCodeHashProcessing prometheus.Histogram
)

// Init module initialization
func Init() {
	metrics = true

	blocks = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "blocks",
		Help:      "The total number of processed blocks",
	})
	transactions = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "transactions",
		Help:      "The total number of processed transactions",
	})
	receipts = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "receipts",
		Help:      "The total number of processed receipts",
	})
	logs = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "logs",
		Help:      "The total number of processed logs",
	})
	accessListEntries = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "access_list_entries",
		Help:      "The total number of processed access list entries",
	})

	lenPayloadChan = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "len_payload_chan",
		Help:      "Current length of publishPayload",
	})

	tPayloadDecode = promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: statsSubsystem,
		Name:      "t_payload_decode",
		Help:      "Payload decoding time",
	})
	tFreePostgres = promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: statsSubsystem,
		Name:      "t_free_postgres",
		Help:      "Time spent waiting for free postgres tx",
	})
	tPostgresCommit = promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: statsSubsystem,
		Name:      "t_postgres_commit",
		Help:      "Postgres transaction commit duration",
	})
	tHeaderProcessing = promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: statsSubsystem,
		Name:      "t_header_processing",
		Help:      "Header processing time",
	})
	tUncleProcessing = promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: statsSubsystem,
		Name:      "t_uncle_processing",
		Help:      "Uncle processing time",
	})
	tTxAndRecProcessing = promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: statsSubsystem,
		Name:      "t_tx_receipt_processing",
		Help:      "Tx and receipt processing time",
	})
	tStateAndStoreProcessing = promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: statsSubsystem,
		Name:      "t_state_store_processing",
		Help:      "State and storage processing time",
	})
	tCodeAndCodeHashProcessing = promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: statsSubsystem,
		Name:      "t_code_codehash_processing",
		Help:      "Code and codehash processing time",
	})
}

// RegisterDBCollector create metric colletor for given connection
func RegisterDBCollector(name string, db *sqlx.DB) {
	if metrics {
		prometheus.Register(NewDBStatsCollector(name, db))
	}
}

// BlockInc block counter increment
func BlockInc() {
	if metrics {
		blocks.Inc()
	}
}

// TransactionInc transaction counter increment
func TransactionInc() {
	if metrics {
		transactions.Inc()
	}
}

// ReceiptInc receipt counter increment
func ReceiptInc() {
	if metrics {
		receipts.Inc()
	}
}

// LogInc log counter increment
func LogInc() {
	if metrics {
		logs.Inc()
	}
}

// AccessListElementInc access list element counter increment
func AccessListElementInc() {
	if metrics {
		accessListEntries.Inc()
	}
}

// SetLenPayloadChan set chan length
func SetLenPayloadChan(ln int) {
	if metrics {
		lenPayloadChan.Set(float64(ln))
	}
}

// SetTimeMetric time metric observation
func SetTimeMetric(name string, t time.Duration) {
	if !metrics {
		return
	}
	tAsF64 := t.Seconds()
	switch name {
	case "t_payload_decode":
		tPayloadDecode.Observe(tAsF64)
	case "t_free_postgres":
		tFreePostgres.Observe(tAsF64)
	case "t_postgres_commit":
		tPostgresCommit.Observe(tAsF64)
	case "t_header_processing":
		tHeaderProcessing.Observe(tAsF64)
	case "t_uncle_processing":
		tUncleProcessing.Observe(tAsF64)
	case "t_tx_receipt_processing":
		tTxAndRecProcessing.Observe(tAsF64)
	case "t_state_store_processing":
		tStateAndStoreProcessing.Observe(tAsF64)
	case "t_code_codehash_processing":
		tCodeAndCodeHashProcessing.Observe(tAsF64)
	}
}
