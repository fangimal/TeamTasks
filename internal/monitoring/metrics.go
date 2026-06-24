package monitoring

import (
	"database/sql"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	namespace = "teamtasks"
	subsystem = "http"
)

var (
	HTTPRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "requests_total",
			Help:      "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	HTTPRequestDurationSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "request_duration_seconds",
			Help:      "Duration of HTTP requests in seconds",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)
)

func init() {
	prometheus.MustRegister(HTTPRequestsTotal)
	prometheus.MustRegister(HTTPRequestDurationSeconds)
}

func Handler() http.Handler {
	return promhttp.Handler()
}

func RegisterDBMetrics(database *sql.DB) {
	prometheus.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "db_max_open_connections",
			Help:      "Maximum number of open connections to the database",
		},
		func() float64 {
			return float64(database.Stats().MaxOpenConnections)
		},
	))

	prometheus.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "db_open_connections",
			Help:      "Current number of open connections",
		},
		func() float64 {
			return float64(database.Stats().OpenConnections)
		},
	))

	prometheus.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "db_in_use_connections",
			Help:      "Current number of in-use connections",
		},
		func() float64 {
			return float64(database.Stats().InUse)
		},
	))

	prometheus.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "db_idle_connections",
			Help:      "Current number of idle connections",
		},
		func() float64 {
			return float64(database.Stats().Idle)
		},
	))

	prometheus.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "db_wait_count_total",
			Help:      "Total number of connections waited for",
		},
		func() float64 {
			return float64(database.Stats().WaitCount)
		},
	))

	prometheus.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "db_wait_duration_seconds_total",
			Help:      "Total time blocked waiting for a new connection",
		},
		func() float64 {
			return database.Stats().WaitDuration.Seconds()
		},
	))
}
