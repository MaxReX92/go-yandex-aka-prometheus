package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/logger"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics"
)

type metricsPusherConfig interface {
	MetricsServerURL() string
	PushMetricsTimeout() time.Duration
}

type httpMetricsPusher struct {
	client           http.Client
	metricsServerURL string
	pushTimeout      time.Duration
}

func NewMetricsPusher(config metricsPusherConfig) MetricsPusher {
	return &httpMetricsPusher{
		client:           http.Client{},
		metricsServerURL: strings.TrimRight(config.MetricsServerURL(), "/"),
		pushTimeout:      config.PushMetricsTimeout(),
	}
}

func (p *httpMetricsPusher) Push(ctx context.Context, metrics []metrics.Metric) error {
	logger.InfoFormat("Push %v metrics", len(metrics))

	pushCtx, cancel := context.WithTimeout(ctx, p.pushTimeout)
	defer cancel()

	for _, metric := range metrics {
		metricType := metric.GetType()
		metricName := metric.GetName()
		metricValue := metric.GetStringValue()
		defer metric.Flush()

		// http://<АДРЕС_СЕРВЕРА>/update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>;
		url := fmt.Sprintf("%v/update/%v/%v/%v", p.metricsServerURL, metricType, metricName, metricValue)
		request, err := http.NewRequestWithContext(pushCtx, http.MethodPost, url, nil)
		if err != nil {
			logger.ErrorFormat("Fail to create push request: %v", err.Error())
			return err
		}
		request.Header.Add("Content-Type", "text/plain")

		response, err := p.client.Do(request)
		if err != nil {
			logger.ErrorFormat("Fail to push metric: %v", err.Error())
			return err
		}
		defer response.Body.Close()

		content, err := io.ReadAll(response.Body)
		if err != nil {
			logger.ErrorFormat("Fail to read response body: %v", err.Error())
			return err
		}

		stringContent := string(content)
		if response.StatusCode != http.StatusOK {
			logger.ErrorFormat("Unexpected response status code: %v %v", response.Status, stringContent)
			return fmt.Errorf("fail to push metric: %v", stringContent)
		}

		logger.InfoFormat("Pushed metric: %v. value: %v, status: %v %v",
			metricName, metric.GetStringValue(), response.Status, stringContent)
	}
	return nil
}
