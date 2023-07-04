package consul

import (
	"context"
	"errors"

	"github.com/bldsoft/geos/pkg/client"
	"github.com/bldsoft/geos/pkg/entity"
	"github.com/bldsoft/geos/pkg/storage"
	"github.com/bldsoft/gost/consul"
	"github.com/hashicorp/consul/api"
)

type discoveredClient struct {
	clientLoader *loader[client.Client]
}

func newDiscoveredClient(serviceCluster string, discovery *consul.Discovery, makeClient func(ci api.AgentServiceChecksInfo) (client.Client, error)) *discoveredClient {
	return &discoveredClient{clientLoader: newLoader[client.Client](
		serviceCluster,
		discovery,
		makeClient,
	)}
}

func (c *discoveredClient) GeoNameContinents(ctx context.Context) []*entity.GeoNameContinent {
	return storage.GeoNameContinents()
}

func doWithClientLoader[C any, R any](loader *loader[C], stopOnErr bool, doRequest func(client C) (res R, err error)) (R, error) {
	var zero R
	clients, err := loader.Load()
	if err != nil {
		return zero, err
	}

	var multiErr error
	for _, client := range clients {
		res, err := doRequest(client)
		if err == nil {
			return res, nil
		}
		multiErr = errors.Join(multiErr, err)
		if stopOnErr {
			break
		}
	}
	return zero, multiErr
}

func (c *discoveredClient) Country(ctx context.Context, address string) (*entity.Country, error) {
	return doWithClientLoader[client.Client, *entity.Country](c.clientLoader, true,
		func(client client.Client) (res *entity.Country, err error) {
			return client.Country(ctx, address)
		})
}

func (c *discoveredClient) City(ctx context.Context, address string) (*entity.City, error) {
	return doWithClientLoader[client.Client, *entity.City](c.clientLoader, true,
		func(client client.Client) (res *entity.City, err error) {
			return client.City(ctx, address)
		})
}

func (c *discoveredClient) CityLite(ctx context.Context, address, lang string) (*entity.CityLite, error) {
	return doWithClientLoader[client.Client, *entity.CityLite](c.clientLoader, true,
		func(client client.Client) (res *entity.CityLite, err error) {
			return client.CityLite(ctx, address, lang)
		})
}

func (c *discoveredClient) GeoNameCountries(ctx context.Context, filter entity.GeoNameFilter) ([]*entity.GeoNameCountry, error) {
	return doWithClientLoader[client.Client, []*entity.GeoNameCountry](c.clientLoader, true,
		func(client client.Client) (res []*entity.GeoNameCountry, err error) {
			return client.GeoNameCountries(ctx, filter)
		})
}

func (c *discoveredClient) GeoNameSubdivisions(ctx context.Context, filter entity.GeoNameFilter) ([]*entity.GeoNameAdminSubdivision, error) {
	return doWithClientLoader[client.Client, []*entity.GeoNameAdminSubdivision](c.clientLoader, true,
		func(client client.Client) (res []*entity.GeoNameAdminSubdivision, err error) {
			return client.GeoNameSubdivisions(ctx, filter)
		})
}

func (c *discoveredClient) GeoNameCities(ctx context.Context, filter entity.GeoNameFilter) ([]*entity.GeoName, error) {
	return doWithClientLoader[client.Client, []*entity.GeoName](c.clientLoader, true,
		func(client client.Client) (res []*entity.GeoName, err error) {
			return client.GeoNameCities(ctx, filter)
		})
}
