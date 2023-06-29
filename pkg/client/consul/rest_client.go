package consul

import (
	"context"
	"fmt"

	"github.com/bldsoft/geos/pkg/client"
	rest_client "github.com/bldsoft/geos/pkg/client/rest"
	"github.com/bldsoft/geos/pkg/config"
	"github.com/bldsoft/geos/pkg/entity"
	"github.com/bldsoft/geos/pkg/microservice"
	"github.com/bldsoft/gost/consul"
	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/consul/api"
)

type restClient struct {
	*discoveredClient
}

func NewRestClient(discovery *consul.Discovery) *restClient {
	c := &discoveredClient{clientLoader: newLoader[client.Client](
		config.ConsulRestClusterName,
		discovery,
		func(info api.AgentServiceChecksInfo) (client.Client, error) {
			c, err := rest_client.NewClient(fmt.Sprintf("%s:%d", info.Service.Address, info.Service.Port))
			if err != nil {
				return nil, err
			}
			return c.SetApiKey(info.Service.Meta[microservice.ConsulAPIMetaKey]), nil
		},
	)}
	return &restClient{c}
}

func (c *restClient) GeoIPDump(ctx context.Context) (*resty.Response, error) {
	return doWithClientLoader[client.Client, *resty.Response](c.clientLoader, true,
		func(client client.Client) (res *resty.Response, err error) {
			return client.(*rest_client.Client).GeoIPDump(ctx)
		})
}

func (c *restClient) GeoNameDump(ctx context.Context, filter entity.GeoNameFilter) (*resty.Response, error) {
	return doWithClientLoader[client.Client, *resty.Response](c.clientLoader, true,
		func(client client.Client) (res *resty.Response, err error) {
			return client.(*rest_client.Client).GeoNameDump(ctx, filter)
		})
}
