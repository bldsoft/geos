// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.15.8
// source: geoip.proto

package __

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

// GeoServiceClient is the client API for GeoService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type GeoServiceClient interface {
	Country(ctx context.Context, in *IpRequest, opts ...grpc.CallOption) (*CountryResponse, error)
	City(ctx context.Context, in *IpRequest, opts ...grpc.CallOption) (*CityResponse, error)
}

type geoServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewGeoServiceClient(cc grpc.ClientConnInterface) GeoServiceClient {
	return &geoServiceClient{cc}
}

func (c *geoServiceClient) Country(ctx context.Context, in *IpRequest, opts ...grpc.CallOption) (*CountryResponse, error) {
	out := new(CountryResponse)
	err := c.cc.Invoke(ctx, "/geoip.GeoService/Country", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *geoServiceClient) City(ctx context.Context, in *IpRequest, opts ...grpc.CallOption) (*CityResponse, error) {
	out := new(CityResponse)
	err := c.cc.Invoke(ctx, "/geoip.GeoService/City", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GeoServiceServer is the server API for GeoService service.
// All implementations must embed UnimplementedGeoServiceServer
// for forward compatibility
type GeoServiceServer interface {
	Country(context.Context, *IpRequest) (*CountryResponse, error)
	City(context.Context, *IpRequest) (*CityResponse, error)
	mustEmbedUnimplementedGeoServiceServer()
}

// UnimplementedGeoServiceServer must be embedded to have forward compatible implementations.
type UnimplementedGeoServiceServer struct {
}

func (UnimplementedGeoServiceServer) Country(context.Context, *IpRequest) (*CountryResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Country not implemented")
}
func (UnimplementedGeoServiceServer) City(context.Context, *IpRequest) (*CityResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method City not implemented")
}
func (UnimplementedGeoServiceServer) mustEmbedUnimplementedGeoServiceServer() {}

// UnsafeGeoServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to GeoServiceServer will
// result in compilation errors.
type UnsafeGeoServiceServer interface {
	mustEmbedUnimplementedGeoServiceServer()
}

func RegisterGeoServiceServer(s grpc.ServiceRegistrar, srv GeoServiceServer) {
	s.RegisterService(&GeoService_ServiceDesc, srv)
}

func _GeoService_Country_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(IpRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GeoServiceServer).Country(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/geoip.GeoService/Country",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GeoServiceServer).Country(ctx, req.(*IpRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _GeoService_City_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(IpRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GeoServiceServer).City(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/geoip.GeoService/City",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GeoServiceServer).City(ctx, req.(*IpRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// GeoService_ServiceDesc is the grpc.ServiceDesc for GeoService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var GeoService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "geoip.GeoService",
	HandlerType: (*GeoServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Country",
			Handler:    _GeoService_Country_Handler,
		},
		{
			MethodName: "City",
			Handler:    _GeoService_City_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "geoip.proto",
}
