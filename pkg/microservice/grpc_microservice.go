package microservice

import (
	context "context"
	"net"

	"github.com/bldsoft/geos/pkg/controller"
	grpc_controller "github.com/bldsoft/geos/pkg/controller/grpc"
	pb "github.com/bldsoft/geos/pkg/controller/grpc/proto"
	"github.com/bldsoft/gost/log"
	grpc "google.golang.org/grpc"
)

type GrpcMicroservice struct {
	address      string
	grpcServer   *grpc.Server
	geoIpService controller.GeoIpService
}

func NewGrpcMicroservice(address string, geoIpService controller.GeoIpService) *GrpcMicroservice {
	return &GrpcMicroservice{
		address:      address,
		geoIpService: geoIpService,
	}
}

func (s *GrpcMicroservice) registerServices() {
	geoIpController := grpc_controller.NewGeoIpController(s.geoIpService)
	pb.RegisterGeoIpServiceServer(s.grpcServer, geoIpController)
}

func (s *GrpcMicroservice) Run() error {
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

func (s *GrpcMicroservice) Stop(ctx context.Context) error {
	stopped := make(chan struct{})
	go func() {
		defer close(stopped)
		s.grpcServer.GracefulStop()
	}()
	select {
	case <-stopped:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
