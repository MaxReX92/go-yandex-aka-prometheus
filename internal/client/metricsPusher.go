package client

import (
	"context"
	"errors"
	"fmt"
	"go-yandex-aka-prometheus/internal/logger"
	"go-yandex-aka-prometheus/internal/metrics"
	"io"
	"net/http"
	"strings"
	"time"
)

type MetricsPusherConfig struct {
	MetricsServerUrl string
	PushTimeout      time.Duration
}

type MetricsPusher struct {
	client           http.Client
	metricsServerUrl string
	pushTimeout      time.Duration
}

func NewMetricsPusher(config MetricsPusherConfig) MetricsPusher {
	return MetricsPusher{
		client:           http.Client{},
		metricsServerUrl: strings.TrimRight(config.MetricsServerUrl, "/"),
		pushTimeout:      config.PushTimeout,
	}
}

func (p *MetricsPusher) Push(ctx context.Context, metrics []metrics.Metric) error {
	logger.InfoFormat("Push %v metrics", len(metrics))

	pushCtx, cancel := context.WithTimeout(ctx, p.pushTimeout)
	defer cancel()

	for _, metric := range metrics {
		metricType := metric.GetType()
		metricName := metric.GetName()
		metricValue := metric.StringValue()

		// http://<АДРЕС_СЕРВЕРА>/update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>;
		url := fmt.Sprintf("%v/update/%v/%v/%v", p.metricsServerUrl, metricType, metricName, metricValue)
		request, err := http.NewRequestWithContext(pushCtx, "POST", url, nil)
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
			return errors.New(fmt.Sprintf("Fail to push metric: %v", content))
		}

		logger.InfoFormat("Pushed metric: %v. value: %v, status: %v %v",
			metricName, metric.StringValue(), response.Status, stringContent)
	}
	return nil
}
