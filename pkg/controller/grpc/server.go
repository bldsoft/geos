package grpc

import (
	context "context"
	"net"

	"github.com/bldsoft/geos/pkg/controller"
	pb "github.com/bldsoft/geos/pkg/controller/grpc/proto"
	"github.com/bldsoft/gost/log"
	grpc "google.golang.org/grpc"
)

//go:generate protoc -I=../../.. --go_out=proto --go-grpc_out=proto api/grpc/geoip.proto

type Server struct {
	address      string
	grpcServer   *grpc.Server
	geoIpService controller.GeoIpService
}

func NewServer(address string, geoIpService controller.GeoIpService) *Server {
	return &Server{
		address:      address,
		geoIpService: geoIpService,
	}
}

func (s *Server) registerServices() {
	geoIpController := NewGeoIpController(s.geoIpService)
	pb.RegisterGeoIpServiceServer(s.grpcServer, geoIpController)
}

func (s *Server) Run() error {
	lis, err := net.Listen("tcp", s.address)
	if err != nil {
		return err
	}
	var opts []grpc.ServerOption

	s.grpcServer = grpc.NewServer(opts...)
	s.registerServices()

	log.Infof("Grpc server started. Listening on %s", s.address)
	defer log.Infof("Grpc server stopped")
	return s.grpcServer.Serve(lis)
}

func (s *Server) Stop(ctx context.Context) error {
	stopped := make(chan struct{})
	go func() {
		close(stopped)
		s.grpcServer.GracefulStop()
	}()
	select {
	case <-stopped:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
