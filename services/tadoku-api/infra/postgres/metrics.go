package postgres

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
)

// PoolCollector exports pgx pool acquisition statistics to Prometheus.
type PoolCollector struct {
	pool                *pgxpool.Pool
	acquireCount        *prometheus.Desc
	acquiredConnections *prometheus.Desc
	emptyAcquireCount   *prometheus.Desc
	acquireDuration     *prometheus.Desc
}

func NewPoolCollector(pool *pgxpool.Pool) *PoolCollector {
	return &PoolCollector{
		pool: pool,
		acquireCount: prometheus.NewDesc(
			"tadoku_api_postgres_pool_acquire_count_total",
			"Total number of successful PostgreSQL pool acquisitions.",
			nil,
			nil,
		),
		acquiredConnections: prometheus.NewDesc(
			"tadoku_api_postgres_pool_acquired_connections",
			"Number of PostgreSQL connections currently acquired from the pool.",
			nil,
			nil,
		),
		emptyAcquireCount: prometheus.NewDesc(
			"tadoku_api_postgres_pool_empty_acquire_count_total",
			"Total number of successful PostgreSQL pool acquisitions that waited for a connection.",
			nil,
			nil,
		),
		acquireDuration: prometheus.NewDesc(
			"tadoku_api_postgres_pool_acquire_duration_seconds_total",
			"Total time spent on successful PostgreSQL pool acquisitions.",
			nil,
			nil,
		),
	}
}

func (c *PoolCollector) Describe(descriptions chan<- *prometheus.Desc) {
	descriptions <- c.acquireCount
	descriptions <- c.acquiredConnections
	descriptions <- c.emptyAcquireCount
	descriptions <- c.acquireDuration
}

func (c *PoolCollector) Collect(metrics chan<- prometheus.Metric) {
	stats := c.pool.Stat()
	metrics <- prometheus.MustNewConstMetric(c.acquireCount, prometheus.CounterValue, float64(stats.AcquireCount()))
	metrics <- prometheus.MustNewConstMetric(c.acquiredConnections, prometheus.GaugeValue, float64(stats.AcquiredConns()))
	metrics <- prometheus.MustNewConstMetric(c.emptyAcquireCount, prometheus.CounterValue, float64(stats.EmptyAcquireCount()))
	metrics <- prometheus.MustNewConstMetric(c.acquireDuration, prometheus.CounterValue, stats.AcquireDuration().Seconds())
}
