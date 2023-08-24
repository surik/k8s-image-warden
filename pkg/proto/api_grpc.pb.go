// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v4.23.4
// source: pkg/proto/api.proto

package proto

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
	ControllerService_Report_FullMethodName    = "/proto.ControllerService/Report"
	ControllerService_GetReport_FullMethodName = "/proto.ControllerService/GetReport"
	ControllerService_GetRules_FullMethodName  = "/proto.ControllerService/GetRules"
	ControllerService_Validate_FullMethodName  = "/proto.ControllerService/Validate"
	ControllerService_Mutate_FullMethodName    = "/proto.ControllerService/Mutate"
)

// ControllerServiceClient is the client API for ControllerService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ControllerServiceClient interface {
	Report(ctx context.Context, in *ReportRequest, opts ...grpc.CallOption) (*ReportResponse, error)
	GetReport(ctx context.Context, in *GetReportRequest, opts ...grpc.CallOption) (*GetReportResponse, error)
	GetRules(ctx context.Context, in *GetRulesRequest, opts ...grpc.CallOption) (*GetRulesResponse, error)
	Validate(ctx context.Context, in *ValidateRequest, opts ...grpc.CallOption) (*ValidateResponse, error)
	Mutate(ctx context.Context, in *MutateRequest, opts ...grpc.CallOption) (*MutateResponse, error)
}

type controllerServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewControllerServiceClient(cc grpc.ClientConnInterface) ControllerServiceClient {
	return &controllerServiceClient{cc}
}

func (c *controllerServiceClient) Report(ctx context.Context, in *ReportRequest, opts ...grpc.CallOption) (*ReportResponse, error) {
	out := new(ReportResponse)
	err := c.cc.Invoke(ctx, ControllerService_Report_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *controllerServiceClient) GetReport(ctx context.Context, in *GetReportRequest, opts ...grpc.CallOption) (*GetReportResponse, error) {
	out := new(GetReportResponse)
	err := c.cc.Invoke(ctx, ControllerService_GetReport_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *controllerServiceClient) GetRules(ctx context.Context, in *GetRulesRequest, opts ...grpc.CallOption) (*GetRulesResponse, error) {
	out := new(GetRulesResponse)
	err := c.cc.Invoke(ctx, ControllerService_GetRules_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *controllerServiceClient) Validate(ctx context.Context, in *ValidateRequest, opts ...grpc.CallOption) (*ValidateResponse, error) {
	out := new(ValidateResponse)
	err := c.cc.Invoke(ctx, ControllerService_Validate_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *controllerServiceClient) Mutate(ctx context.Context, in *MutateRequest, opts ...grpc.CallOption) (*MutateResponse, error) {
	out := new(MutateResponse)
	err := c.cc.Invoke(ctx, ControllerService_Mutate_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ControllerServiceServer is the server API for ControllerService service.
// All implementations must embed UnimplementedControllerServiceServer
// for forward compatibility
type ControllerServiceServer interface {
	Report(context.Context, *ReportRequest) (*ReportResponse, error)
	GetReport(context.Context, *GetReportRequest) (*GetReportResponse, error)
	GetRules(context.Context, *GetRulesRequest) (*GetRulesResponse, error)
	Validate(context.Context, *ValidateRequest) (*ValidateResponse, error)
	Mutate(context.Context, *MutateRequest) (*MutateResponse, error)
	mustEmbedUnimplementedControllerServiceServer()
}

// UnimplementedControllerServiceServer must be embedded to have forward compatible implementations.
type UnimplementedControllerServiceServer struct {
}

func (UnimplementedControllerServiceServer) Report(context.Context, *ReportRequest) (*ReportResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Report not implemented")
}
func (UnimplementedControllerServiceServer) GetReport(context.Context, *GetReportRequest) (*GetReportResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetReport not implemented")
}
func (UnimplementedControllerServiceServer) GetRules(context.Context, *GetRulesRequest) (*GetRulesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetRules not implemented")
}
func (UnimplementedControllerServiceServer) Validate(context.Context, *ValidateRequest) (*ValidateResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Validate not implemented")
}
func (UnimplementedControllerServiceServer) Mutate(context.Context, *MutateRequest) (*MutateResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Mutate not implemented")
}
func (UnimplementedControllerServiceServer) mustEmbedUnimplementedControllerServiceServer() {}

// UnsafeControllerServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ControllerServiceServer will
// result in compilation errors.
type UnsafeControllerServiceServer interface {
	mustEmbedUnimplementedControllerServiceServer()
}

func RegisterControllerServiceServer(s grpc.ServiceRegistrar, srv ControllerServiceServer) {
	s.RegisterService(&ControllerService_ServiceDesc, srv)
}

func _ControllerService_Report_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ReportRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ControllerServiceServer).Report(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ControllerService_Report_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ControllerServiceServer).Report(ctx, req.(*ReportRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ControllerService_GetReport_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetReportRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ControllerServiceServer).GetReport(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ControllerService_GetReport_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ControllerServiceServer).GetReport(ctx, req.(*GetReportRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ControllerService_GetRules_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetRulesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ControllerServiceServer).GetRules(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ControllerService_GetRules_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ControllerServiceServer).GetRules(ctx, req.(*GetRulesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ControllerService_Validate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ValidateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ControllerServiceServer).Validate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ControllerService_Validate_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ControllerServiceServer).Validate(ctx, req.(*ValidateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ControllerService_Mutate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MutateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ControllerServiceServer).Mutate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ControllerService_Mutate_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ControllerServiceServer).Mutate(ctx, req.(*MutateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// ControllerService_ServiceDesc is the grpc.ServiceDesc for ControllerService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ControllerService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "proto.ControllerService",
	HandlerType: (*ControllerServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Report",
			Handler:    _ControllerService_Report_Handler,
		},
		{
			MethodName: "GetReport",
			Handler:    _ControllerService_GetReport_Handler,
		},
		{
			MethodName: "GetRules",
			Handler:    _ControllerService_GetRules_Handler,
		},
		{
			MethodName: "Validate",
			Handler:    _ControllerService_Validate_Handler,
		},
		{
			MethodName: "Mutate",
			Handler:    _ControllerService_Mutate_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "pkg/proto/api.proto",
}
