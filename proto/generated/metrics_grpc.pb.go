// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v4.23.3
// source: proto/metrics.proto

package generated

import (
	context "context"

	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	MetricServer_GetValue_FullMethodName     = "/com.github.MaxReX92.go_yandex_aka_prometheus.MetricServer/GetValue"
	MetricServer_UpdateValues_FullMethodName = "/com.github.MaxReX92.go_yandex_aka_prometheus.MetricServer/UpdateValues"
	MetricServer_Ping_FullMethodName         = "/com.github.MaxReX92.go_yandex_aka_prometheus.MetricServer/Ping"
	MetricServer_Report_FullMethodName       = "/com.github.MaxReX92.go_yandex_aka_prometheus.MetricServer/Report"
)

// MetricServerClient is the client API for MetricServer service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type MetricServerClient interface {
	GetValue(ctx context.Context, in *MetricsRequest, opts ...grpc.CallOption) (*MetricsResponse, error)
	UpdateValues(ctx context.Context, in *MetricsRequest, opts ...grpc.CallOption) (*MetricsResponse, error)
	Ping(ctx context.Context, in *Nothing, opts ...grpc.CallOption) (*Response, error)
	Report(ctx context.Context, in *Nothing, opts ...grpc.CallOption) (*ReportResponse, error)
}

type metricServerClient struct {
	cc grpc.ClientConnInterface
}

func NewMetricServerClient(cc grpc.ClientConnInterface) MetricServerClient {
	return &metricServerClient{cc}
}

func (c *metricServerClient) GetValue(ctx context.Context, in *MetricsRequest, opts ...grpc.CallOption) (*MetricsResponse, error) {
	out := new(MetricsResponse)
	err := c.cc.Invoke(ctx, MetricServer_GetValue_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *metricServerClient) UpdateValues(ctx context.Context, in *MetricsRequest, opts ...grpc.CallOption) (*MetricsResponse, error) {
	out := new(MetricsResponse)
	err := c.cc.Invoke(ctx, MetricServer_UpdateValues_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *metricServerClient) Ping(ctx context.Context, in *Nothing, opts ...grpc.CallOption) (*Response, error) {
	out := new(Response)
	err := c.cc.Invoke(ctx, MetricServer_Ping_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *metricServerClient) Report(ctx context.Context, in *Nothing, opts ...grpc.CallOption) (*ReportResponse, error) {
	out := new(ReportResponse)
	err := c.cc.Invoke(ctx, MetricServer_Report_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MetricServerServer is the server API for MetricServer service.
// All implementations must embed UnimplementedMetricServerServer
// for forward compatibility
type MetricServerServer interface {
	GetValue(context.Context, *MetricsRequest) (*MetricsResponse, error)
	UpdateValues(context.Context, *MetricsRequest) (*MetricsResponse, error)
	Ping(context.Context, *Nothing) (*Response, error)
	Report(context.Context, *Nothing) (*ReportResponse, error)
	mustEmbedUnimplementedMetricServerServer()
}

// UnimplementedMetricServerServer must be embedded to have forward compatible implementations.
type UnimplementedMetricServerServer struct {
}

func (UnimplementedMetricServerServer) GetValue(context.Context, *MetricsRequest) (*MetricsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetValue not implemented")
}
func (UnimplementedMetricServerServer) UpdateValues(context.Context, *MetricsRequest) (*MetricsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateValues not implemented")
}
func (UnimplementedMetricServerServer) Ping(context.Context, *Nothing) (*Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Ping not implemented")
}
func (UnimplementedMetricServerServer) Report(context.Context, *Nothing) (*ReportResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Report not implemented")
}
func (UnimplementedMetricServerServer) mustEmbedUnimplementedMetricServerServer() {}

// UnsafeMetricServerServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to MetricServerServer will
// result in compilation errors.
type UnsafeMetricServerServer interface {
	mustEmbedUnimplementedMetricServerServer()
}

func RegisterMetricServerServer(s grpc.ServiceRegistrar, srv MetricServerServer) {
	s.RegisterService(&MetricServer_ServiceDesc, srv)
}

func _MetricServer_GetValue_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MetricsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MetricServerServer).GetValue(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MetricServer_GetValue_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MetricServerServer).GetValue(ctx, req.(*MetricsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MetricServer_UpdateValues_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MetricsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MetricServerServer).UpdateValues(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MetricServer_UpdateValues_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MetricServerServer).UpdateValues(ctx, req.(*MetricsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MetricServer_Ping_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Nothing)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MetricServerServer).Ping(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MetricServer_Ping_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MetricServerServer).Ping(ctx, req.(*Nothing))
	}
	return interceptor(ctx, in, info, handler)
}

func _MetricServer_Report_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Nothing)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MetricServerServer).Report(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MetricServer_Report_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MetricServerServer).Report(ctx, req.(*Nothing))
	}
	return interceptor(ctx, in, info, handler)
}

// MetricServer_ServiceDesc is the grpc.ServiceDesc for MetricServer service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var MetricServer_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "com.github.MaxReX92.go_yandex_aka_prometheus.MetricServer",
	HandlerType: (*MetricServerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetValue",
			Handler:    _MetricServer_GetValue_Handler,
		},
		{
			MethodName: "UpdateValues",
			Handler:    _MetricServer_UpdateValues_Handler,
		},
		{
			MethodName: "Ping",
			Handler:    _MetricServer_Ping_Handler,
		},
		{
			MethodName: "Report",
			Handler:    _MetricServer_Report_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/metrics.proto",
}
