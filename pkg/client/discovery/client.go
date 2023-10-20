package discovery

import (
	"errors"

	"github.com/bldsoft/geos/pkg/client"
	grpc_client "github.com/bldsoft/geos/pkg/client/grpc"
	"github.com/bldsoft/geos/pkg/config"
	"github.com/bldsoft/geos/pkg/microservice"
	"github.com/bldsoft/gost/discovery"
)

var ErrServiceNotFound = errors.New("service not found")
var ErrGRPCDisabled = errors.New("grpc is disabled")

func NewClient(d discovery.Discovery) client.Client {
	return &client.MultiClient{Clients: []client.Client{NewGrpcClient(d), NewRestClient(d)}}
}

func NewGrpcClient(d discovery.Discovery) client.Client {
	return &discoveredClient{
		clientLoader: newLoader[client.Client](
			config.ServiceName,
			d,
			func(info discovery.ServiceInstanceInfo) (client.Client, error) {
				if grpcAddr := info.Meta[microservice.GrpcAddressMetaKey]; grpcAddr != "" {
					return grpc_client.NewClient(grpcAddr)
				}
				return nil, ErrGRPCDisabled
			},
		)}
}
