package worker

import "github.com/prometheus/client_golang/prometheus"

type Metrics struct {
	InFlight                *prometheus.GaugeVec
	Attempts                *prometheus.CounterVec
	Duration                *prometheus.HistogramVec
	Pending                 *prometheus.GaugeVec
	Failed                  *prometheus.GaugeVec
	OldestDueAge            *prometheus.GaugeVec
	Unsupported             prometheus.Gauge
	UnsupportedRunning      prometheus.Gauge
	UnsupportedFailed       prometheus.Gauge
	UnsupportedOldestDueAge prometheus.Gauge
	ExpiredLeases           *prometheus.CounterVec
}

func NewMetrics(registry *prometheus.Registry) *Metrics {
	metrics := &Metrics{
		InFlight: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "tadoku_worker_in_flight",
			Help: "Active async tasks by predefined type.",
		}, []string{"type"}),
		Attempts: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "tadoku_worker_failed_attempts_total",
			Help: "Failed handler attempts by predefined type and failure code.",
		}, []string{"type", "code"}),
		Duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "tadoku_worker_handler_duration_seconds",
			Help:    "Handler and task transition duration by predefined type.",
			Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1, 5, 15, 30},
		}, []string{"type"}),
		Pending: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "tadoku_worker_pending_tasks",
			Help: "Pending async tasks by predefined type.",
		}, []string{"type"}),
		Failed: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "tadoku_worker_failed_tasks",
			Help: "Terminally failed async tasks by predefined type.",
		}, []string{"type"}),
		OldestDueAge: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "tadoku_worker_oldest_due_age_seconds",
			Help: "Age of the oldest due or expired async task by predefined type.",
		}, []string{"type"}),
		Unsupported: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "tadoku_worker_unsupported_pending_tasks",
			Help: "Pending tasks with an unsupported type; they remain unclaimed.",
		}),
		UnsupportedRunning: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "tadoku_worker_unsupported_running_tasks",
			Help: "Running jobs with a type unsupported by this executable.",
		}),
		UnsupportedFailed: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "tadoku_worker_unsupported_failed_tasks",
			Help: "Failed jobs with a type unsupported by this executable.",
		}),
		UnsupportedOldestDueAge: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "tadoku_worker_unsupported_oldest_due_age_seconds",
			Help: "Age of the oldest due pending task with an unsupported type.",
		}),
		ExpiredLeases: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "tadoku_worker_expired_leases_total",
			Help: "Expired running tasks reclaimed by predefined type.",
		}, []string{"type"}),
	}
	registry.MustRegister(metrics.InFlight, metrics.Attempts, metrics.Duration, metrics.Pending, metrics.Failed, metrics.OldestDueAge, metrics.Unsupported, metrics.UnsupportedRunning, metrics.UnsupportedFailed, metrics.UnsupportedOldestDueAge, metrics.ExpiredLeases)
	metrics.Unsupported.Set(0)
	metrics.UnsupportedOldestDueAge.Set(0)
	return metrics
}

func (m *Metrics) initialize(handlers *registry) {
	for _, entry := range handlers.ordered {
		spec := entry.spec
		m.InFlight.WithLabelValues(string(spec.typeName)).Set(0)
		m.Duration.WithLabelValues(string(spec.typeName))
		m.Pending.WithLabelValues(string(spec.typeName)).Set(0)
		m.Failed.WithLabelValues(string(spec.typeName)).Set(0)
		m.OldestDueAge.WithLabelValues(string(spec.typeName)).Set(0)
		m.ExpiredLeases.WithLabelValues(string(spec.typeName))
	}
}
