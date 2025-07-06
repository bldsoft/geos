package discovery

import (
	"context"
	"errors"

	"github.com/bldsoft/geos/pkg/client"
	"github.com/bldsoft/geos/pkg/entity"
	"github.com/bldsoft/geos/pkg/storage/geonames"
	"github.com/bldsoft/geos/pkg/storage/state"
	"github.com/bldsoft/gost/discovery"
)

type discoveredClient struct {
	clientLoader *loader[client.Client]
}

func newDiscoveredClient(serviceCluster string, discovery discovery.Discovery, makeClient func(serviceInfo discovery.ServiceInstanceInfo) (client.Client, error)) *discoveredClient {
	return &discoveredClient{clientLoader: newLoader[client.Client](
		serviceCluster,
		discovery,
		makeClient,
	)}
}

func (c *discoveredClient) GeoNameContinents(ctx context.Context) []*entity.GeoNameContinent {
	return geonames.GeoNameContinents()
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

func (c *discoveredClient) City(ctx context.Context, address string, includeISP bool) (*entity.City, error) {
	return doWithClientLoader[client.Client, *entity.City](c.clientLoader, true,
		func(client client.Client) (res *entity.City, err error) {
			return client.City(ctx, address, includeISP)
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

func (c *discoveredClient) CheckUpdates(ctx context.Context) (entity.Updates, error) {
	return doWithClientLoader(c.clientLoader, true,
		func(client client.Client) (res entity.Updates, err error) {
			return client.CheckUpdates(ctx)
		})
}

func (c *discoveredClient) Update(ctx context.Context) (entity.Updates, error) {
	return doWithClientLoader(c.clientLoader, true,
		func(client client.Client) (res entity.Updates, err error) {
			return client.Update(ctx)
		})
}

func (c *discoveredClient) State(ctx context.Context) (*state.GeosState, error) {
	return doWithClientLoader(c.clientLoader, true,
		func(client client.Client) (res *state.GeosState, err error) {
			return client.State(ctx)
		})
}
