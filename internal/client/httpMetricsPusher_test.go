package client

import (
	"context"
	"github.com/stretchr/testify/assert"
	"go-yandex-aka-prometheus/internal/metrics"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type config struct {
	connetionString string
	timeout         time.Duration
}

func TestHttpMetricsPusher_Push(t *testing.T) {
	tests := []struct {
		name          string
		metricsToPush []metrics.Metric
		expectedURLs  map[string]bool
	}{
		{
			name:          "empty_metrics_list",
			metricsToPush: []metrics.Metric{},
			expectedURLs:  map[string]bool{},
		},
		{
			name: "simple_metrics",
			metricsToPush: []metrics.Metric{
				createCounterMetric("counterMetric1", 100),
				createGaugeMetric("gaugeMetric1", 100.001),
			},
			expectedURLs: map[string]bool{
				"/update/counter/counterMetric1/100": false,
				"/update/gauge/gaugeMetric1/100.001": false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, ok := tt.expectedURLs[r.URL.Path]
				assert.Truef(t, ok, "Unexpected url call: %v", r.URL)
				assert.Equal(t, "text/plain", r.Header.Get("Content-Type"))
				tt.expectedURLs[r.URL.Path] = true
			}))
			defer server.Close()

			pusher := NewMetricsPusher(&config{
				connetionString: server.URL,
				timeout:         10 * time.Second,
			})

			err := pusher.Push(ctx, tt.metricsToPush)
			assert.NoError(t, err)

			for url, called := range tt.expectedURLs {
				assert.True(t, called, "Url %v was not called", url)
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
	return c.connetionString
}

func (c *config) PushMetricsTimeout() time.Duration {
	return c.timeout
}
