package discovery

import (
	"errors"

	"github.com/bldsoft/geos/pkg/client"
	grpc_client "github.com/bldsoft/geos/pkg/client/grpc"
	"github.com/bldsoft/geos/pkg/config"
	"github.com/bldsoft/gost/discovery"
)

var ErrServiceNotFound = errors.New("service not found")

func NewClient(d discovery.Discovery) client.Client {
	return &client.MultiClient{Clients: []client.Client{NewGrpcClient(d), NewRestClient(d)}}
}

func NewGrpcClient(d discovery.Discovery) client.Client {
	return &discoveredClient{clientLoader: newLoader[client.Client](
		config.GrpcServiceName,
		d,
		func(info discovery.ServiceInstanceInfo) (client.Client, error) {
			return grpc_client.NewClient(info.HostPort())
		},
	)}
}
