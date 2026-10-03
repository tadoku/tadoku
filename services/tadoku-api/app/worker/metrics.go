package worker

import "github.com/prometheus/client_golang/prometheus"

type Metrics struct {
	inFlight                *prometheus.GaugeVec
	attempts                *prometheus.CounterVec
	duration                *prometheus.HistogramVec
	pending                 *prometheus.GaugeVec
	failed                  *prometheus.GaugeVec
	oldestDueAge            *prometheus.GaugeVec
	unsupported             prometheus.Gauge
	unsupportedRunning      prometheus.Gauge
	unsupportedFailed       prometheus.Gauge
	unsupportedOldestDueAge prometheus.Gauge
	expiredLeases           *prometheus.CounterVec
}

func NewMetrics(registry *prometheus.Registry) *Metrics {
	metrics := &Metrics{
		inFlight: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "tadoku_worker_in_flight",
			Help: "Active jobs by predefined type.",
		}, []string{"type"}),
		attempts: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "tadoku_worker_failed_attempts_total",
			Help: "Failed handler attempts by predefined type and failure code.",
		}, []string{"type", "code", "tenant_kind"}),
		duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "tadoku_worker_handler_duration_seconds",
			Help:    "Handler execution duration before the job transition by predefined type.",
			Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1, 5, 15, 30},
		}, []string{"type", "tenant_kind"}),
		pending: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "tadoku_worker_pending_jobs",
			Help: "Pending jobs by predefined type.",
		}, []string{"type"}),
		failed: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "tadoku_worker_failed_jobs",
			Help: "Terminally failed jobs by predefined type.",
		}, []string{"type"}),
		oldestDueAge: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "tadoku_worker_oldest_due_age_seconds",
			Help: "Age of the oldest due or expired job by predefined type.",
		}, []string{"type"}),
		unsupported: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "tadoku_worker_unsupported_pending_jobs",
			Help: "Pending jobs with an unsupported type; they remain unclaimed.",
		}),
		unsupportedRunning: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "tadoku_worker_unsupported_running_jobs",
			Help: "Running jobs with a type unsupported by this executable.",
		}),
		unsupportedFailed: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "tadoku_worker_unsupported_failed_jobs",
			Help: "Failed jobs with a type unsupported by this executable.",
		}),
		unsupportedOldestDueAge: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "tadoku_worker_unsupported_oldest_due_age_seconds",
			Help: "Age of the oldest due pending job with an unsupported type.",
		}),
		expiredLeases: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "tadoku_worker_expired_leases_total",
			Help: "Expired running jobs reclaimed by predefined type.",
		}, []string{"type"}),
	}
	registry.MustRegister(
		metrics.inFlight,
		metrics.attempts,
		metrics.duration,
		metrics.pending,
		metrics.failed,
		metrics.oldestDueAge,
		metrics.unsupported,
		metrics.unsupportedRunning,
		metrics.unsupportedFailed,
		metrics.unsupportedOldestDueAge,
		metrics.expiredLeases,
	)
	metrics.unsupported.Set(0)
	metrics.unsupportedOldestDueAge.Set(0)
	return metrics
}

func (m *Metrics) initialize(handlers *registry) {
	for _, entry := range handlers.ordered {
		spec := entry.spec
		m.inFlight.WithLabelValues(string(spec.typeName)).Set(0)
		for _, kind := range []string{"production", "test", "unknown"} {
			m.duration.WithLabelValues(string(spec.typeName), kind)
		}
		m.pending.WithLabelValues(string(spec.typeName)).Set(0)
		m.failed.WithLabelValues(string(spec.typeName)).Set(0)
		m.oldestDueAge.WithLabelValues(string(spec.typeName)).Set(0)
		m.expiredLeases.WithLabelValues(string(spec.typeName))
	}
}
