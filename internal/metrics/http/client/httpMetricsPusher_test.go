package client

import (
	"context"
	"encoding/json"
	"hash"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	internalHash "github.com/MaxReX92/go-yandex-aka-prometheus/internal/hash"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics"
	metricsHttp "github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics/http"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics/model"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics/types"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/parser"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/test"
)

type testConf struct {
	connectionString string
	key              []byte
	timeout          time.Duration
	parallelLimit    int
	signEnabled      bool
}

type testMetric struct {
	name       string
	metricType string
	hash       []byte
	value      float64
}

func TestHttpMetricsPusher_Push(t *testing.T) {
	var (
		counterValue int64 = 100
		gaugeValue         = 100.001
		lock               = sync.Mutex{}
	)

	tests := []struct {
		name                 string
		expectedErrorMessage string
		metricsToPush        []metrics.Metric
		expectedRequests     []model.Metrics
		responseStatusCode   int
	}{
		{
			name:               "empty_metrics_list",
			metricsToPush:      []metrics.Metric{},
			expectedRequests:   []model.Metrics{},
			responseStatusCode: http.StatusOK,
		},
		{
			name: "unknown_metric_type",
			metricsToPush: []metrics.Metric{
				&testMetric{metricType: "invalid_type"},
			},
			expectedErrorMessage: "failed to convert metric with type invalid_type",
		},
		{
			name: "wrong_status_code",
			metricsToPush: []metrics.Metric{
				createCounterMetric("counterMetric1", counterValue),
			},
			expectedErrorMessage: "failed to push metric: ",
			responseStatusCode:   http.StatusBadRequest,
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
			responseStatusCode: http.StatusOK,
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
				modelRequest := []*model.Metrics{}
				err := json.NewDecoder(r.Body).Decode(&modelRequest)
				assert.NoError(t, err)
				for _, modelMetric := range modelRequest {
					lock.Lock()
					defer lock.Unlock()

					called[modelMetric.ID+modelMetric.MType] = true
				}

				w.WriteHeader(tt.responseStatusCode)
			}))
			defer server.Close()

			conf := &testConf{
				connectionString: server.URL,
				timeout:          10 * time.Second,
				signEnabled:      false,
				key:              nil,
				parallelLimit:    10,
			}
			signer := internalHash.NewSigner(conf)
			converter := metricsHttp.NewMetricsConverter(conf, signer)
			pusher, err := NewPusher(conf, converter, nil)
			assert.NoError(t, err)

			err = pusher.Push(ctx, test.ArrayToChan(tt.metricsToPush))

			if tt.expectedErrorMessage != "" {
				assert.ErrorContains(t, err, tt.expectedErrorMessage)
			}

			for key, call := range called {
				assert.True(t, call, "Metric was not pushed, %v", key)
			}
		})
	}
}

func Test_URLNormalization(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedError string
		expectedURL   string
	}{
		{
			name:          "empty_url",
			input:         "",
			expectedError: "failed to normalize url: empty url string",
		},
		{
			name:        "no_schema_no_port",
			input:       "127.0.0.1",
			expectedURL: "http://127.0.0.1",
		},
		{
			name:        "no_schema_port",
			input:       "127.0.0.1:1234",
			expectedURL: "http://127.0.0.1:1234",
		},
		{
			name:        "schema_port",
			input:       "ftp://127.0.0.1:1234",
			expectedURL: "ftp://127.0.0.1:1234",
		},
		{
			name:        "localhost",
			input:       "localhost:1234",
			expectedURL: "http://localhost:1234",
		},
		{
			name:        "valid",
			input:       "https://ya.ru",
			expectedURL: "https://ya.ru",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := normalizeURL(tt.input)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err.Error())
			} else {
				assert.Equal(t, tt.expectedURL, actual.String())
			}
		})
	}
}

func createCounterMetric(name string, value int64) metrics.Metric {
	metric := types.NewCounterMetric(name)
	metric.SetValue(float64(value))
	return metric
}

func createGaugeMetric(name string, value float64) metrics.Metric {
	metric := types.NewGaugeMetric(name)
	metric.SetValue(value)
	return metric
}

func (c *testConf) MetricsServerURL() string {
	return c.connectionString
}

func (c *testConf) PushMetricsTimeout() time.Duration {
	return c.timeout
}

func (t *testMetric) GetName() string {
	return t.name
}

func (t *testMetric) GetType() string {
	return t.metricType
}

func (t *testMetric) GetValue() float64 {
	return t.value
}

func (t *testMetric) GetStringValue() string {
	return parser.FloatToString(t.value)
}

func (t *testMetric) SetValue(value float64) float64 {
	t.value = value
	return value
}

func (t *testMetric) Flush() {
}

func (t *testMetric) GetHash(hash.Hash) ([]byte, error) {
	return t.hash, nil
}

func (c *testConf) SignMetrics() bool {
	return c.signEnabled
}

func (c *testConf) GetKey() []byte {
	return c.key
}

func (c *testConf) ParallelLimit() int {
	return c.parallelLimit
}
