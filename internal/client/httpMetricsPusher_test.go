package client

import (
	"context"
	"encoding/json"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/model"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics"
)

type config struct {
	connectionString string
	timeout          time.Duration
}

func TestHttpMetricsPusher_Push(t *testing.T) {
	var counterValue int64 = 100
	var gaugeValue float64 = 100.001

	tests := []struct {
		name             string
		metricsToPush    []metrics.Metric
		expectedRequests []model.Metrics
	}{
		{
			name:             "empty_metrics_list",
			metricsToPush:    []metrics.Metric{},
			expectedRequests: []model.Metrics{},
		},
		{
			name: "simple_metrics",
			metricsToPush: []metrics.Metric{
				createCounterMetric("counterMetric1", counterValue),
				createGaugeMetric("gaugeMetric1", gaugeValue),
			},
			expectedRequests: []model.Metrics{
				{
					ID:    "counterMetric1",
					MType: "counter",
					Delta: &counterValue,
				}, {
					ID:    "gaugeMetric1",
					MType: "gauge",
					Value: &gaugeValue,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			called := map[string]bool{}
			for _, request := range tt.expectedRequests {
				called[request.ID+request.MType] = false
			}

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				defer r.Body.Close()
				modelRequest := &model.Metrics{}
				err := json.NewDecoder(r.Body).Decode(modelRequest)
				assert.NoError(t, err)
				called[modelRequest.ID+modelRequest.MType] = true
			}))
			defer server.Close()

			pusher := NewMetricsPusher(&config{
				connectionString: server.URL,
				timeout:          10 * time.Second,
			})

			err := pusher.Push(ctx, tt.metricsToPush)
			assert.NoError(t, err)

			for key, call := range called {
				assert.True(t, call, "Metric was not pushed, %v", key)
			}
		})
	}

}

func createCounterMetric(name string, value int64) metrics.Metric {
	metric := metrics.NewCounterMetric(name)
	metric.SetValue(float64(value))
	return metric
}

func createGaugeMetric(name string, value float64) metrics.Metric {
	metric := metrics.NewGaugeMetric(name)
	metric.SetValue(value)
	return metric
}

func (c *config) MetricsServerURL() string {
	return c.connectionString
}

func (c *config) PushMetricsTimeout() time.Duration {
	return c.timeout
}
