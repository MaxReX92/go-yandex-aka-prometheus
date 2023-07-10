package client

import (
	"context"
	"fmt"

	rpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/logger"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics/grpc"
	"github.com/MaxReX92/go-yandex-aka-prometheus/pkg/chunk"
	"github.com/MaxReX92/go-yandex-aka-prometheus/proto/generated"
)

const chunkSize int = 10

type GrpcMetricsPusherConfig interface {
	GrpcServerURL() string
}

type grpcMetricsPusher struct {
	client    generated.MetricServerClient
	converter *grpc.Converter
}

func NewPusher(conf GrpcMetricsPusherConfig, converter *grpc.Converter) (*grpcMetricsPusher, error) {
	connection, err := rpc.Dial(conf.GrpcServerURL(), rpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, logger.WrapError("open grpc connection", err)
	}

	return &grpcMetricsPusher{
		client:    generated.NewMetricServerClient(connection),
		converter: converter,
	}, nil
}

func (g *grpcMetricsPusher) Push(ctx context.Context, metricsChan <-chan metrics.Metric) error {
	for _, metricsChunk := range chunk.ChanToChunks(metricsChan, chunkSize) {
		chunkLen := len(metricsChunk)
		requestMetrics := make([]*generated.Metric, chunkLen)

		for i := 0; i < chunkLen; i++ {
			requestMetric, err := g.converter.ToModelMetric(metricsChunk[i])
			if err != nil {
				return logger.WrapError("generate update metric request", err)
			}

			requestMetrics[i] = requestMetric
		}

		request := &generated.MetricsRequest{
			Metrics: requestMetrics,
		}

		response, err := g.client.UpdateValues(ctx, request)
		if err != nil {
			return logger.WrapError("call update metrics procedure", err)
		}

		if response.Status != generated.Status_OK {
			logger.ErrorFormat("Unexpected response status code: %s %s", response.Status, *response.Error)
			return logger.WrapError(fmt.Sprintf("push metrics: %s %s", response.Status, *response.Error), metrics.ErrUnexpectedStatusCode)
		}

		for _, metric := range metricsChunk {
			logger.InfoFormat("Pushed metric: %v. value: %v, status: %v", metric.GetName(), metric.GetStringValue(), response.Status)
			metric.Flush()
		}
	}

	return nil
}
