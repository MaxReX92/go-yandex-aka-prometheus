package internal

import (
	"context"
	"errors"
	"fmt"
	"go-yandex-aka-prometheus/internal/metrics"
	"io"
	"log"
	"net/http"
	"strings"
)

type MetricsPusherConfig struct {
	MetricsServerUrl string
}

type MetricsPusher struct {
	client           http.Client
	metricsServerUrl string
}

func NewMetricsPusher(config MetricsPusherConfig) MetricsPusher {
	return MetricsPusher{
		client:           http.Client{},
		metricsServerUrl: strings.TrimRight(config.MetricsServerUrl, "/"),
	}
}

func (p *MetricsPusher) Push(ctx context.Context, metrics []metrics.Metric) error {
	pushCtx, cancel := context.WithCancel(ctx) // TODO: timeout
	defer cancel()

	for _, metric := range metrics {

		// http://<АДРЕС_СЕРВЕРА>/update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>;
		url := fmt.Sprintf("%v/update/%v/%v/%v", p.metricsServerUrl, metric.GetType(), metric.GetName(), metric.StringValue())

		request, err := http.NewRequestWithContext(pushCtx, "POST", url, nil)
		if err != nil {
			return err
		}

		response, err := p.client.Do(request)
		defer response.Body.Close()
		if err != nil {
			return err
		}

		content, err := io.ReadAll(response.Body)
		if err != nil {
			return err
		}

		if response.StatusCode != http.StatusOK {
			return errors.New(fmt.Sprintf("Fail to push metric:%v", string(content)))
		}

		log.Printf("Metric %v - OK")
	}
	return nil
}
