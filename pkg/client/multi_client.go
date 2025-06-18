package client

import (
	"context"
	"errors"

	"github.com/bldsoft/geos/pkg/entity"
	"github.com/hashicorp/go-multierror"
)

type MultiClient struct {
	Clients []Client
}

func getFromAny[T any](ctx context.Context, clients []Client, f func(ctx context.Context, client Client) (*T, error)) (*T, error) {
	if len(clients) == 0 {
		return nil, errors.New("no clients")
	}
	var multiErr error
	for _, client := range clients {
		obj, err := f(ctx, client)
		if err == nil {
			return obj, nil
		}
		multiErr = multierror.Append(multiErr, err)
	}
	return nil, multiErr
}

func (c *MultiClient) Country(ctx context.Context, address string) (*entity.Country, error) {
	return getFromAny(ctx, c.Clients, func(ctx context.Context, client Client) (*entity.Country, error) {
		return client.Country(ctx, address)
	})
}

func (c *MultiClient) City(ctx context.Context, address string, includeISP bool) (*entity.City, error) {
	return getFromAny(ctx, c.Clients, func(ctx context.Context, client Client) (*entity.City, error) {
		return client.City(ctx, address, includeISP)
	})
}

func (c *MultiClient) CityLite(ctx context.Context, address, lang string) (*entity.CityLite, error) {
	return getFromAny(ctx, c.Clients, func(ctx context.Context, client Client) (*entity.CityLite, error) {
		return client.CityLite(ctx, address, lang)
	})
}

func getManyFromAny[T any](ctx context.Context, clients []Client, f func(ctx context.Context, client Client) ([]*T, error)) ([]*T, error) {
	var multiErr error
	for _, client := range clients {
		obj, err := f(ctx, client)
		if err == nil {
			return obj, nil
		}
		multiErr = multierror.Append(multiErr, err)
	}
	return nil, multiErr
}

func (c *MultiClient) GeoNameContinents(ctx context.Context) []*entity.GeoNameContinent {
	continents, _ := getManyFromAny(ctx, c.Clients, func(ctx context.Context, client Client) ([]*entity.GeoNameContinent, error) {
		return client.GeoNameContinents(ctx), nil
	})
	return continents
}

func (c *MultiClient) GeoNameCountries(ctx context.Context, filter entity.GeoNameFilter) ([]*entity.GeoNameCountry, error) {
	return getManyFromAny(ctx, c.Clients, func(ctx context.Context, client Client) ([]*entity.GeoNameCountry, error) {
		return client.GeoNameCountries(ctx, filter)
	})
}
func (c *MultiClient) GeoNameSubdivisions(ctx context.Context, filter entity.GeoNameFilter) ([]*entity.GeoNameAdminSubdivision, error) {
	return getManyFromAny(ctx, c.Clients, func(ctx context.Context, client Client) ([]*entity.GeoNameAdminSubdivision, error) {
		return client.GeoNameSubdivisions(ctx, filter)
	})
}
func (c *MultiClient) GeoNameCities(ctx context.Context, filter entity.GeoNameFilter) ([]*entity.GeoName, error) {
	return getManyFromAny(ctx, c.Clients, func(ctx context.Context, client Client) ([]*entity.GeoName, error) {
		return client.GeoNameCities(ctx, filter)
	})
}

func (c *MultiClient) CheckUpdates(ctx context.Context) (entity.Updates, error) {
	return nil, nil
}

func (c *MultiClient) Update(ctx context.Context) (entity.Updates, error) {
	return nil, nil
}
