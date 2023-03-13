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
	"github.com/prometheus/client_golang/prometheus"

	dbmetrics "github.com/ethereum/go-ethereum/statediff/indexer/database/metrics"
)

const (
	namespace = "eth_statediff_service"
	subsystem = "connections"
)

// DBStatsGetter is an interface that gets sql.DBStats.
type DBStatsGetter interface {
	DbStats() dbmetrics.DbStats
}

// DBStatsCollector implements the prometheus.Collector interface.
type DBStatsCollector struct {
	sg DBStatsGetter

	// descriptions of exported metrics
	maxOpenDesc           *prometheus.Desc
	openDesc              *prometheus.Desc
	inUseDesc             *prometheus.Desc
	idleDesc              *prometheus.Desc
	waitedForDesc         *prometheus.Desc
	blockedSecondsDesc    *prometheus.Desc
	closedMaxIdleDesc     *prometheus.Desc
	closedMaxLifetimeDesc *prometheus.Desc
}

// NewDBStatsCollector creates a new DBStatsCollector.
func NewDBStatsCollector(dbName string, sg DBStatsGetter) *DBStatsCollector {
	labels := prometheus.Labels{"db_name": dbName}
	return &DBStatsCollector{
		sg: sg,
		maxOpenDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "max_open"),
			"Maximum number of open connections to the database.",
			nil,
			labels,
		),
		openDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "open"),
			"The number of established connections both in use and idle.",
			nil,
			labels,
		),
		inUseDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "in_use"),
			"The number of connections currently in use.",
			nil,
			labels,
		),
		idleDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "idle"),
			"The number of idle connections.",
			nil,
			labels,
		),
		waitedForDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "waited_for"),
			"The total number of connections waited for.",
			nil,
			labels,
		),
		blockedSecondsDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "blocked_seconds"),
			"The total time blocked waiting for a new connection.",
			nil,
			labels,
		),
		closedMaxIdleDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "closed_max_idle"),
			"The total number of connections closed due to SetMaxIdleConns.",
			nil,
			labels,
		),
		closedMaxLifetimeDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "closed_max_lifetime"),
			"The total number of connections closed due to SetConnMaxLifetime.",
			nil,
			labels,
		),
	}
}

// Describe implements the prometheus.Collector interface.
func (c DBStatsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.maxOpenDesc
	ch <- c.openDesc
	ch <- c.inUseDesc
	ch <- c.idleDesc
	ch <- c.waitedForDesc
	ch <- c.blockedSecondsDesc
	ch <- c.closedMaxIdleDesc
	ch <- c.closedMaxLifetimeDesc
}

// Collect implements the prometheus.Collector interface.
func (c DBStatsCollector) Collect(ch chan<- prometheus.Metric) {
	stats := c.sg.DbStats()

	ch <- prometheus.MustNewConstMetric(
		c.maxOpenDesc,
		prometheus.GaugeValue,
		float64(stats.MaxOpen()),
	)
	ch <- prometheus.MustNewConstMetric(
		c.openDesc,
		prometheus.GaugeValue,
		float64(stats.Open()),
	)
	ch <- prometheus.MustNewConstMetric(
		c.inUseDesc,
		prometheus.GaugeValue,
		float64(stats.InUse()),
	)
	ch <- prometheus.MustNewConstMetric(
		c.idleDesc,
		prometheus.GaugeValue,
		float64(stats.Idle()),
	)
	ch <- prometheus.MustNewConstMetric(
		c.waitedForDesc,
		prometheus.CounterValue,
		float64(stats.WaitCount()),
	)
	ch <- prometheus.MustNewConstMetric(
		c.blockedSecondsDesc,
		prometheus.CounterValue,
		stats.WaitDuration().Seconds(),
	)
	ch <- prometheus.MustNewConstMetric(
		c.closedMaxIdleDesc,
		prometheus.CounterValue,
		float64(stats.MaxIdleClosed()),
	)
	ch <- prometheus.MustNewConstMetric(
		c.closedMaxLifetimeDesc,
		prometheus.CounterValue,
		float64(stats.MaxLifetimeClosed()),
	)
}
