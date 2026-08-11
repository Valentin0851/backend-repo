package observability

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

func TestRecapCacheMetrics(t *testing.T) {
	metrics := NewRecapCacheMetrics()
	metrics.Hit()
	metrics.Miss()
	metrics.Error("get")

	registry := prometheus.NewRegistry()
	registry.MustRegister(metrics.Collectors()...)
	families, err := registry.Gather()
	if err != nil {
		t.Fatalf("gather metrics: %v", err)
	}

	requests := metricFamily(t, families, "gooffer_recap_cache_requests_total")
	if counterValue(requests, "result", "hit") != 1 {
		t.Fatal("hit counter was not incremented")
	}
	if counterValue(requests, "result", "miss") != 1 {
		t.Fatal("miss counter was not incremented")
	}
	errors := metricFamily(t, families, "gooffer_recap_cache_errors_total")
	if counterValue(errors, "operation", "get") != 1 {
		t.Fatal("get error counter was not incremented")
	}
}

func TestPostgresPoolCollectors(t *testing.T) {
	pool, err := pgxpool.New(
		context.Background(),
		"postgres://user:password@127.0.0.1:1/database?connect_timeout=1",
	)
	if err != nil {
		t.Fatalf("create pool: %v", err)
	}
	defer pool.Close()

	registry := prometheus.NewRegistry()
	registry.MustRegister(NewPostgresPoolCollectors(pool)...)
	families, err := registry.Gather()
	if err != nil {
		t.Fatalf("gather metrics: %v", err)
	}

	connections := metricFamily(t, families, "gooffer_postgres_pool_connections")
	if len(connections.Metric) != 4 {
		t.Fatalf("connection states = %d, want 4", len(connections.Metric))
	}
	metricFamily(t, families, "gooffer_postgres_pool_acquires_total")
	metricFamily(t, families, "gooffer_postgres_pool_acquire_duration_seconds_total")
	metricFamily(t, families, "gooffer_postgres_pool_canceled_acquires_total")
	metricFamily(t, families, "gooffer_postgres_pool_new_connections_total")
}

func metricFamily(t *testing.T, families []*dto.MetricFamily, name string) *dto.MetricFamily {
	t.Helper()
	for _, family := range families {
		if family.GetName() == name {
			return family
		}
	}
	t.Fatalf("metric family %q not found", name)
	return nil
}

func counterValue(family *dto.MetricFamily, labelName, labelValue string) float64 {
	for _, metric := range family.Metric {
		for _, label := range metric.Label {
			if label.GetName() == labelName && label.GetValue() == labelValue {
				return metric.GetCounter().GetValue()
			}
		}
	}
	return 0
}
