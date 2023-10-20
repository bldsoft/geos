package discovery

import (
	"context"

	"github.com/bldsoft/geos/pkg/client"
	rest_client "github.com/bldsoft/geos/pkg/client/rest"
	"github.com/bldsoft/geos/pkg/config"
	"github.com/bldsoft/geos/pkg/entity"
	"github.com/bldsoft/geos/pkg/microservice"
	"github.com/bldsoft/gost/discovery"
	"github.com/go-resty/resty/v2"
)

type restClient struct {
	*discoveredClient
}

func NewRestClient(d discovery.Discovery) *restClient {
	c := &discoveredClient{clientLoader: newLoader[client.Client](
		config.ServiceName,
		d,
		func(serviceInfo discovery.ServiceInstanceInfo) (client.Client, error) {
			c, err := rest_client.NewClient(string(serviceInfo.Address))
			if err != nil {
				return nil, err
			}
			return c.SetApiKey(serviceInfo.Meta[microservice.APIKeyMetaKey]), nil
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
