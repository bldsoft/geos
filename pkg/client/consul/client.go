package consul

import (
	"errors"
	"fmt"

	"github.com/bldsoft/geos/pkg/client"
	grpc_client "github.com/bldsoft/geos/pkg/client/grpc"
	"github.com/bldsoft/geos/pkg/config"
	"github.com/bldsoft/gost/consul"
	"github.com/hashicorp/consul/api"
)

var ErrServiceNotFound = errors.New("service not found")

func NewClientFromDiscovery(d *consul.Discovery) client.Client {
	return &client.MultiClient{Clients: []client.Client{NewGrpcClient(d), NewRestClient(d)}}
}

func NewGrpcClient(discovery *consul.Discovery) client.Client {
	return &discoveredClient{clientLoader: newLoader[client.Client](
		config.ConsulGrpcClusterName,
		discovery,
		func(info api.AgentServiceChecksInfo) (client.Client, error) {
			return grpc_client.NewClient(fmt.Sprintf("%s:%d", info.Service.Address, info.Service.Port))
		},
	)}
}
