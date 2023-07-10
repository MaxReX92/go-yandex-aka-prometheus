package server

import (
	"context"
	"net"

	rpc "google.golang.org/grpc"

	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/logger"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics/grpc"
	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics/server"
	"github.com/MaxReX92/go-yandex-aka-prometheus/proto/generated"
)

type GrpcServerConfig interface {
	ListenTCP() string
}

type grpcServer struct {
	generated.UnimplementedMetricServerServer

	listenTCP      string
	converter      *grpc.Converter
	requestHandler server.RequestHandler
	server         *rpc.Server
}

func New(conf GrpcServerConfig, converter *grpc.Converter, requestHandler server.RequestHandler) *grpcServer {
	return &grpcServer{
		listenTCP:      conf.ListenTCP(),
		converter:      converter,
		requestHandler: requestHandler,
		server:         rpc.NewServer(),
	}
}

func (g *grpcServer) Start(_ context.Context) error {
	listen, err := net.Listen("tcp", g.listenTCP)
	if err != nil {
		return logger.WrapError("start listen TCP", err)
	}
	generated.RegisterMetricServerServer(g.server, g)

	logger.Info("Start gRPC service")
	err = g.server.Serve(listen)
	if err != nil {
		return logger.WrapError("start gRPC service", err)
	}

	return nil
}

func (g *grpcServer) Stop(ctx context.Context) error {
	logger.Info("Stopping gRPC service")
	g.server.GracefulStop()
	return nil
}

func (g *grpcServer) GetValue(ctx context.Context, request *generated.MetricsRequest) (*generated.MetricsResponse, error) {
	if request.Metrics == nil {
		logger.Error("failed to get metric value: invalid request")
		return g.createMetricResponse(generated.Status_ERROR, nil, "invalid request"), nil
	}

	metricsCount := len(request.Metrics)
	responseMetrics := make([]*generated.Metric, metricsCount)
	for i := 0; i < metricsCount; i++ {
		metric, err := g.converter.FromModelMetric(request.Metrics[i])
		if err != nil {
			return g.createMetricResponse(generated.Status_ERROR, nil, logger.WrapError("convert metric request", err).Error()), nil
		}

		result, err := g.requestHandler.GetMetricValue(ctx, metric.GetType(), metric.GetName())
		if err != nil {
			return g.createMetricResponse(generated.Status_ERROR, nil, logger.WrapError("get metric value", err).Error()), nil
		}

		response, err := g.converter.ToModelMetric(result)
		if err != nil {
			return g.createMetricResponse(generated.Status_ERROR, nil, logger.WrapError("generate response", err).Error()), nil
		}

		responseMetrics[i] = response
	}

	return g.createMetricResponse(generated.Status_OK, responseMetrics, ""), nil
}

func (g *grpcServer) UpdateValues(ctx context.Context, request *generated.MetricsRequest) (*generated.MetricsResponse, error) {
	if request.Metrics == nil {
		logger.Error("failed to get metric value: invalid request")
		return g.createMetricResponse(generated.Status_ERROR, nil, "invalid request"), nil
	}

	metricsCount := len(request.Metrics)
	requestMetrics := make([]metrics.Metric, metricsCount)
	for i := 0; i < metricsCount; i++ {
		metric, err := g.converter.FromModelMetric(request.Metrics[i])
		if err != nil {
			return g.createMetricResponse(generated.Status_ERROR, nil, logger.WrapError("convert metric request", err).Error()), nil
		}

		requestMetrics[i] = metric
	}

	resultMetrics, err := g.requestHandler.UpdateMetricValues(ctx, requestMetrics)
	if err != nil {
		return g.createMetricResponse(generated.Status_ERROR, nil, logger.WrapError("update metrics", err).Error()), nil
	}

	responseMetrics := make([]*generated.Metric, metricsCount)
	for i := 0; i < metricsCount; i++ {
		response, err := g.converter.ToModelMetric(resultMetrics[i])
		if err != nil {
			return g.createMetricResponse(generated.Status_ERROR, nil, logger.WrapError("generate response", err).Error()), nil
		}

		responseMetrics[i] = response
	}

	return g.createMetricResponse(generated.Status_OK, responseMetrics, ""), nil
}

func (g *grpcServer) Ping(ctx context.Context, _ *generated.Nothing) (*generated.Response, error) {
	err := g.requestHandler.Ping(ctx)
	if err != nil {
		return g.createResponse(generated.Status_ERROR, logger.WrapError("ping database", err).Error()), nil
	}

	return g.createResponse(generated.Status_OK, ""), nil
}

func (g *grpcServer) Report(ctx context.Context, _ *generated.Nothing) (*generated.ReportResponse, error) {
	report, err := g.requestHandler.GetReportPage(ctx)
	if err != nil {
		return g.createReportResponse(generated.Status_ERROR, "", logger.WrapError("ping database", err).Error()), nil
	}

	return g.createReportResponse(generated.Status_OK, report, ""), nil
}

func (g *grpcServer) createMetricResponse(status generated.Status, metricsResult []*generated.Metric, errorMessage string) *generated.MetricsResponse {
	response := &generated.MetricsResponse{
		Status: status,
		Result: metricsResult,
	}
	if errorMessage != "" {
		response.Error = &errorMessage
	}

	return response
}

func (g *grpcServer) createResponse(status generated.Status, errorMessage string) *generated.Response {
	response := &generated.Response{
		Status: status,
	}
	if errorMessage != "" {
		response.Error = &errorMessage
	}

	return response
}

func (g *grpcServer) createReportResponse(status generated.Status, report string, errorMessage string) *generated.ReportResponse {
	response := &generated.ReportResponse{
		Status: status,
	}
	if report != "" {
		response.Report = &report
	}
	if errorMessage != "" {
		response.Error = &errorMessage
	}

	return response
}
