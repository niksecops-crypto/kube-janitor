package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds all Prometheus instruments for kube-janitor.
type Metrics struct {
	PodsDeleted    prometheus.Counter
	JobsDeleted    prometheus.Counter
	Errors         prometheus.Counter
	CleanupSeconds prometheus.Histogram
}

// New registers and returns all kube-janitor metrics against the given registerer.
// Pass prometheus.DefaultRegisterer in production and a fresh registry in tests.
func New(reg prometheus.Registerer) *Metrics {
	factory := promauto.With(reg)
	return &Metrics{
		PodsDeleted: factory.NewCounter(prometheus.CounterOpts{
			Namespace: "kube_janitor",
			Name:      "pods_deleted_total",
			Help:      "Total number of finished pods deleted by kube-janitor.",
		}),
		JobsDeleted: factory.NewCounter(prometheus.CounterOpts{
			Namespace: "kube_janitor",
			Name:      "jobs_deleted_total",
			Help:      "Total number of completed/failed jobs deleted by kube-janitor.",
		}),
		Errors: factory.NewCounter(prometheus.CounterOpts{
			Namespace: "kube_janitor",
			Name:      "errors_total",
			Help:      "Total number of errors encountered during cleanup cycles.",
		}),
		CleanupSeconds: factory.NewHistogram(prometheus.HistogramOpts{
			Namespace: "kube_janitor",
			Name:      "cleanup_duration_seconds",
			Help:      "Wall-clock time of each cleanup cycle in seconds.",
			Buckets:   prometheus.DefBuckets,
		}),
	}
}
