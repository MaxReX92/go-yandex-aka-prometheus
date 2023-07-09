package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/crypto"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/logger"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics/model"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics/pusher"
)

type metricsPusherConfig interface {
	ParallelLimit() int
	MetricsServerURL() string
	PushMetricsTimeout() time.Duration
}

type httpMetricsPusher struct {
	converter        *model.MetricsConverter
	client           http.Client
	encryptor        crypto.Encryptor
	metricsServerURL string
	clientIP         string
	parallelLimit    int
	pushTimeout      time.Duration
}

// NewMetricsPusher create new instance of http metrics pusher.
func NewMetricsPusher(config metricsPusherConfig, converter *model.MetricsConverter, encryptor crypto.Encryptor) (pusher.MetricsPusher, error) {
	serverURL, err := normalizeURL(config.MetricsServerURL())
	if err != nil {
		return nil, logger.WrapError("normalize url", err)
	}
	clientIP, err := getClientIP(serverURL)
	if err != nil {
		return nil, logger.WrapError("get client ip", err)
	}

	return &httpMetricsPusher{
		parallelLimit:    config.ParallelLimit(),
		client:           http.Client{},
		encryptor:        encryptor,
		metricsServerURL: serverURL.String(),
		clientIP:         clientIP.String(),
		pushTimeout:      config.PushMetricsTimeout(),
		converter:        converter,
	}, nil
}

func (p *httpMetricsPusher) Push(ctx context.Context, metricsChan <-chan metrics.Metric) error {
	eg, ctx := errgroup.WithContext(ctx)

	for i := 0; i < p.parallelLimit; i++ {
		eg.Go(func() error {
			for {
				select {
				case metric, ok := <-metricsChan:
					if !ok {
						return nil
					}

					err := p.pushMetrics(ctx, []metrics.Metric{metric})
					if err != nil {
						return err
					}
				case <-ctx.Done():
					return ctx.Err()
				}
			}
		})
	}

	return eg.Wait()
}

func (p *httpMetricsPusher) pushMetrics(ctx context.Context, metricsList []metrics.Metric) error {
	metricsCount := len(metricsList)
	if metricsCount == 0 {
		logger.Info("Nothing to push")
	}
	logger.InfoFormat("Push %v metrics", metricsCount)

	pushCtx, cancel := context.WithTimeout(ctx, p.pushTimeout)
	defer cancel()

	modelMetrics := make([]*model.Metrics, metricsCount)
	for i, metric := range metricsList {
		modelMetric, err := p.converter.ToModelMetric(metric)
		if err != nil {
			return logger.WrapError("create model request", err)
		}

		modelMetrics[i] = modelMetric
	}

	buffer := &bytes.Buffer{}
	err := json.NewEncoder(buffer).Encode(modelMetrics)
	if err != nil {
		return logger.WrapError("serialize model request", err)
	}

	if p.encryptor != nil {
		encrypted, err := p.encryptor.Encrypt(buffer.Bytes())
		if err != nil {
			return logger.WrapError("encrypt message", err)
		}

		buffer = bytes.NewBuffer(encrypted)
	}

	request, err := http.NewRequestWithContext(pushCtx, http.MethodPost, p.metricsServerURL+"/updates", buffer)
	if err != nil {
		return logger.WrapError("create push request", err)
	}
	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("X-Real-IP", p.clientIP)

	response, err := p.client.Do(request)
	if err != nil {
		return logger.WrapError("push metrics", err)
	}
	defer response.Body.Close()

	content, err := io.ReadAll(response.Body)
	if err != nil {
		return logger.WrapError("read response body", err)
	}

	stringContent := string(content)
	if response.StatusCode != http.StatusOK {
		logger.ErrorFormat("Unexpected response status code: %v %v", response.Status, stringContent)
		return logger.WrapError(fmt.Sprintf("push metric: %s", stringContent), metrics.ErrUnexpectedStatusCode)
	}

	for _, metric := range metricsList {
		logger.InfoFormat("Pushed metric: %v. value: %v, status: %v", metric.GetName(), metric.GetStringValue(), response.Status)
		metric.Flush()
	}

	return nil
}

func normalizeURL(urlStr string) (*url.URL, error) {
	if urlStr == "" {
		return nil, logger.WrapError("normalize url", metrics.ErrEmptyURL)
	}

	result, err := url.ParseRequestURI(urlStr)
	if err != nil {
		result, err = url.ParseRequestURI("http://" + urlStr)
		if err != nil {
			return nil, logger.WrapError("parse request url", err)
		}
	}

	if result.Scheme == "localhost" {
		// =)
		return normalizeURL("http://" + result.String())
	}

	if result.Scheme == "" {
		result.Scheme = "http"
	}

	return result, nil
}

func getClientIP(serverUrl *url.URL) (net.IP, error) {
	conn, err := net.Dial("udp", fmt.Sprintf("%s:80", serverUrl.Hostname()))
	if err != nil {
		return nil, logger.WrapError("open udp connection with server", err)
	}
	defer conn.Close()

	return conn.LocalAddr().(*net.UDPAddr).IP, nil
}
