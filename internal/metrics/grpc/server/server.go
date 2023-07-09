package server

import (
	"context"

	"github.com/MaxReX92/go-yandex-aka-prometheus/internal/metrics/server"
	"github.com/MaxReX92/go-yandex-aka-prometheus/proto/generated"
)

type grpcServer struct {
	generated.UnimplementedMetricServerServer

	requestHandler server.RequestHandler
}

func NewServer(requestHandler server.RequestHandler) *grpcServer {
	return &grpcServer{
		requestHandler: requestHandler,
	}
}

func (g *grpcServer) GetValue(ctx context.Context, request *generated.MetricsRequest) (*generated.MetricsResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (g *grpcServer) UpdateValues(ctx context.Context, request *generated.MetricsRequest) (*generated.MetricsResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (g *grpcServer) Ping(ctx context.Context, nothing *generated.Nothing) (*generated.Response, error) {
	//TODO implement me
	panic("implement me")
}

func (g *grpcServer) Report(ctx context.Context, nothing *generated.Nothing) (*generated.ReportResponse, error) {
	//TODO implement me
	panic("implement me")
}
