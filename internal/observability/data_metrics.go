package observability

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
)

type RecapCacheMetrics struct {
	requests *prometheus.CounterVec
	errors   *prometheus.CounterVec
}

func NewRecapCacheMetrics() *RecapCacheMetrics {
	metrics := &RecapCacheMetrics{
		requests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "gooffer",
			Subsystem: "recap_cache",
			Name:      "requests_total",
			Help:      "Total number of recap cache lookups by result.",
		}, []string{"result"}),
		errors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "gooffer",
			Subsystem: "recap_cache",
			Name:      "errors_total",
			Help:      "Total number of recap cache errors by operation.",
		}, []string{"operation"}),
	}
	metrics.requests.WithLabelValues("hit").Add(0)
	metrics.requests.WithLabelValues("miss").Add(0)
	for _, operation := range []string{"get", "set", "delete"} {
		metrics.errors.WithLabelValues(operation).Add(0)
	}
	return metrics
}

func (m *RecapCacheMetrics) Hit() {
	m.requests.WithLabelValues("hit").Inc()
}

func (m *RecapCacheMetrics) Miss() {
	m.requests.WithLabelValues("miss").Inc()
}

func (m *RecapCacheMetrics) Error(operation string) {
	m.errors.WithLabelValues(operation).Inc()
}

func (m *RecapCacheMetrics) Collectors() []prometheus.Collector {
	return []prometheus.Collector{m.requests, m.errors}
}

func NewPostgresPoolCollectors(pool *pgxpool.Pool) []prometheus.Collector {
	connectionGauge := func(name, help, state string, value func(*pgxpool.Stat) float64) prometheus.Collector {
		return prometheus.NewGaugeFunc(prometheus.GaugeOpts{
			Namespace:   "gooffer",
			Subsystem:   "postgres_pool",
			Name:        name,
			Help:        help,
			ConstLabels: prometheus.Labels{"state": state},
		}, func() float64 {
			return value(pool.Stat())
		})
	}

	return []prometheus.Collector{
		connectionGauge(
			"connections",
			"Current number of PostgreSQL pool connections by state.",
			"acquired",
			func(stat *pgxpool.Stat) float64 { return float64(stat.AcquiredConns()) },
		),
		connectionGauge(
			"connections",
			"Current number of PostgreSQL pool connections by state.",
			"idle",
			func(stat *pgxpool.Stat) float64 { return float64(stat.IdleConns()) },
		),
		connectionGauge(
			"connections",
			"Current number of PostgreSQL pool connections by state.",
			"total",
			func(stat *pgxpool.Stat) float64 { return float64(stat.TotalConns()) },
		),
		connectionGauge(
			"connections",
			"Current number of PostgreSQL pool connections by state.",
			"max",
			func(stat *pgxpool.Stat) float64 { return float64(stat.MaxConns()) },
		),
		prometheus.NewCounterFunc(prometheus.CounterOpts{
			Namespace: "gooffer",
			Subsystem: "postgres_pool",
			Name:      "acquires_total",
			Help:      "Cumulative number of PostgreSQL pool acquire attempts.",
		}, func() float64 {
			return float64(pool.Stat().AcquireCount())
		}),
		prometheus.NewCounterFunc(prometheus.CounterOpts{
			Namespace: "gooffer",
			Subsystem: "postgres_pool",
			Name:      "acquire_duration_seconds_total",
			Help:      "Cumulative time spent waiting to acquire PostgreSQL connections.",
		}, func() float64 {
			return pool.Stat().AcquireDuration().Seconds()
		}),
		prometheus.NewCounterFunc(prometheus.CounterOpts{
			Namespace: "gooffer",
			Subsystem: "postgres_pool",
			Name:      "canceled_acquires_total",
			Help:      "Cumulative number of canceled PostgreSQL pool acquire attempts.",
		}, func() float64 {
			return float64(pool.Stat().CanceledAcquireCount())
		}),
		prometheus.NewCounterFunc(prometheus.CounterOpts{
			Namespace: "gooffer",
			Subsystem: "postgres_pool",
			Name:      "new_connections_total",
			Help:      "Cumulative number of connections created by the PostgreSQL pool.",
		}, func() float64 {
			return float64(pool.Stat().NewConnsCount())
		}),
	}
}
