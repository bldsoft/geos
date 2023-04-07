package consul

import (
	"fmt"

	"github.com/bldsoft/geos/pkg/client"
	grpc_client "github.com/bldsoft/geos/pkg/client/grpc"
	rest_client "github.com/bldsoft/geos/pkg/client/rest"
	"github.com/bldsoft/geos/pkg/config"
	"github.com/bldsoft/gost/consul"
	"google.golang.org/grpc"
)

func NewClientFromDiscovery(d *consul.Discovery) (client.GeoIPClient, error) {
	if client, err := NewGrpcClient(d); err == nil {
		return client, nil
	}
	return NewRestClient(d)
}

func NewRestClient(d *consul.Discovery) (client.GeoIPClient, error) {
	return rest_client.NewWithClient(config.ConsulRestClusterName, consul.NewHttpClientFromDiscovery(d))
}

func NewGrpcClient(d *consul.Discovery) (client.GeoIPClient, error) {
	_, checks, err := d.ApiClient().Agent().AgentHealthServiceByName(config.ConsulGrpcClusterName)
	if err != nil {
		return nil, err
	}
	if len(checks) == 0 {
		return nil, fmt.Errorf("grpc geos not found")
	}

	return grpc_client.NewClient(
		config.ConsulGrpcClusterName,
		grpc.WithContextDialer(consul.GrpcDialerFromDiscovery(d)),
	)
}
