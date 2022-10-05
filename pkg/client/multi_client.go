package client

import (
	"context"

	"github.com/bldsoft/geos/pkg/entity"
	"github.com/hashicorp/go-multierror"
)

type MultiClient struct {
	Clients []GeoIPClient
}

func getFromAny[T any](ctx context.Context, clients []GeoIPClient, f func(ctx context.Context, client GeoIPClient) (*T, error)) (*T, error) {
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
	return getFromAny(ctx, c.Clients, func(ctx context.Context, client GeoIPClient) (*entity.Country, error) {
		return client.Country(ctx, address)
	})
}

func (c *MultiClient) City(ctx context.Context, address string) (*entity.City, error) {
	return getFromAny(ctx, c.Clients, func(ctx context.Context, client GeoIPClient) (*entity.City, error) {
		return client.City(ctx, address)
	})
}

func (c *MultiClient) CityLite(ctx context.Context, address, lang string) (*entity.CityLite, error) {
	return getFromAny(ctx, c.Clients, func(ctx context.Context, client GeoIPClient) (*entity.CityLite, error) {
		return client.CityLite(ctx, address, lang)
	})
}
