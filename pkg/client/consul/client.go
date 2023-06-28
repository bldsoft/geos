package consul

import (
	"errors"
	"fmt"

	"github.com/bldsoft/geos/pkg/client"
	grpc_client "github.com/bldsoft/geos/pkg/client/grpc"
	rest_client "github.com/bldsoft/geos/pkg/client/rest"
	"github.com/bldsoft/geos/pkg/config"
	"github.com/bldsoft/geos/pkg/microservice"
	"github.com/bldsoft/gost/consul"
	"google.golang.org/grpc"
)

var ErrServiceNotFound = errors.New("service not found")

func NewClientFromDiscovery(d *consul.Discovery) (client.Client, error) {
	res := &client.MultiClient{}
	var multiErr error

	if client, err := NewGrpcClient(d); err == nil {
		res.Clients = append(res.Clients, client)
	} else {
		multiErr = errors.Join(multiErr, err)
	}

	if client, err := NewRestClient(d); err == nil {
		res.Clients = append(res.Clients, client)
	} else {
		multiErr = errors.Join(multiErr, err)
	}

	if len(res.Clients) == 0 {
		return nil, multiErr
	}
	return res, nil
}

func NewRestClient(d *consul.Discovery) (client.Client, error) {
	_, infos, err := d.ApiClient().Agent().AgentHealthServiceByName(config.ConsulRestClusterName)
	if err != nil {
		return nil, err
	}
	if len(infos) == 0 {
		return nil, fmt.Errorf("rest: %w", ErrServiceNotFound)
	}

	c, err := rest_client.NewWithClient(config.ConsulRestClusterName, consul.NewHttpClientFromDiscovery(d))
	if err != nil {
		return nil, err
	}
	return c.SetApiKey(infos[0].Service.Meta[microservice.ConsulAPIMetaKey]), nil
}

func NewGrpcClient(d *consul.Discovery) (client.Client, error) {
	_, checks, err := d.ApiClient().Agent().AgentHealthServiceByName(config.ConsulGrpcClusterName)
	if err != nil {
		return nil, err
	}
	if len(checks) == 0 {
		return nil, fmt.Errorf("grpc: %w", ErrServiceNotFound)
	}

	return grpc_client.NewClient(
		config.ConsulGrpcClusterName,
		grpc.WithContextDialer(consul.GrpcDialerFromDiscovery(d)),
	)
}
