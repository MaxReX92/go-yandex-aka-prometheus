package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/model"
	"io"
	"net/http"
	"net/url"
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

func NewMetricsPusher(config metricsPusherConfig) (MetricsPusher, error) {
	serverURL, err := normalizeURL(config.MetricsServerURL())
	if err != nil {
		return nil, err
	}

	return &httpMetricsPusher{
		client:           http.Client{},
		metricsServerURL: serverURL.String(),
		pushTimeout:      config.PushMetricsTimeout(),
	}, nil
}

func (p *httpMetricsPusher) Push(ctx context.Context, metrics []metrics.Metric) error {
	logger.InfoFormat("Push %v metrics", len(metrics))

	pushCtx, cancel := context.WithTimeout(ctx, p.pushTimeout)
	defer cancel()

	for _, metric := range metrics {

		metricName := metric.GetName()
		modelRequest, err := createModelRequest(metric)
		if err != nil {
			logger.ErrorFormat("Fail to create model request: %v", err.Error())
			return err
		}

		var buffer bytes.Buffer
		err = json.NewEncoder(&buffer).Encode(modelRequest)
		if err != nil {
			logger.ErrorFormat("Fail to serialize model request: %v", err.Error())
			return err
		}

		request, err := http.NewRequestWithContext(pushCtx, http.MethodPost, p.metricsServerURL+"/update", &buffer)
		if err != nil {
			logger.ErrorFormat("Fail to create push request: %v", err.Error())
			return err
		}
		request.Header.Add("Content-Type", "application/json")

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

		metric.Flush()
		logger.InfoFormat("Pushed metric: %v. value: %v, status: %v %v",
			metricName, metric.GetStringValue(), response.Status, stringContent)
	}
	return nil
}

func createModelRequest(metric metrics.Metric) (*model.Metrics, error) {
	modelRequest := &model.Metrics{
		ID:    metric.GetName(),
		MType: metric.GetType(),
	}

	metricValue := metric.GetValue()
	if modelRequest.MType == "counter" {
		counterValue := int64(metricValue)
		modelRequest.Delta = &counterValue
	} else if modelRequest.MType == "gauge" {
		modelRequest.Value = &metricValue
	} else {
		return nil, fmt.Errorf("unknown metric type: %v", modelRequest.MType)
	}

	return modelRequest, nil
}

func normalizeURL(urlStr string) (*url.URL, error) {
	if urlStr == "" {
		return nil, errors.New("empty url string")
	}

	result, err := url.ParseRequestURI(urlStr)
	if err != nil {
		result, err = url.ParseRequestURI("http://" + urlStr)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}
