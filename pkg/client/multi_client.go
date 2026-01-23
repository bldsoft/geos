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

func getFromAny[T any](ctx context.Context, clients []Client, f func(ctx context.Context, client Client) (T, error)) (T, error) {
	if len(clients) == 0 {
		var zero T
		return zero, errors.New("no clients")
	}
	var multiErr error
	for _, client := range clients {
		obj, err := f(ctx, client)
		if err == nil {
			return obj, nil
		}
		multiErr = multierror.Append(multiErr, err)
	}
	var zero T
	return zero, multiErr
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

func getManyFromAny[T any](ctx context.Context, clients []Client, f func(ctx context.Context, client Client) ([]T, error)) ([]T, error) {
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

func (c *MultiClient) CheckGeoIPCityUpdates(ctx context.Context) (entity.DBUpdate[entity.PatchedMMDBVersion], error) {
	return getFromAny(ctx, c.Clients, func(ctx context.Context, client Client) (entity.DBUpdate[entity.PatchedMMDBVersion], error) {
		return client.CheckGeoIPCityUpdates(ctx)
	})
}

func (c *MultiClient) CheckGeoIPISPUpdates(ctx context.Context) (entity.DBUpdate[entity.PatchedMMDBVersion], error) {
	return getFromAny(ctx, c.Clients, func(ctx context.Context, client Client) (entity.DBUpdate[entity.PatchedMMDBVersion], error) {
		return client.CheckGeoIPISPUpdates(ctx)
	})
}

func (c *MultiClient) CheckGeonamesUpdates(ctx context.Context) (entity.DBUpdate[entity.PatchedGeoNamesVersion], error) {
	return getFromAny(ctx, c.Clients, func(ctx context.Context, client Client) (entity.DBUpdate[entity.PatchedGeoNamesVersion], error) {
		return client.CheckGeonamesUpdates(ctx)
	})
}

func (c *MultiClient) UpdateGeoIPCity(ctx context.Context) error {
	_, err := getFromAny(ctx, c.Clients, func(ctx context.Context, client Client) (any, error) {
		return nil, client.UpdateGeoIPCity(ctx)
	})
	return err
}

func (c *MultiClient) UpdateGeoIPISP(ctx context.Context) error {
	_, err := getFromAny(ctx, c.Clients, func(ctx context.Context, client Client) (any, error) {
		return nil, client.UpdateGeoIPISP(ctx)
	})
	return err
}

func (c *MultiClient) UpdateGeonames(ctx context.Context) error {
	_, err := getFromAny(ctx, c.Clients, func(ctx context.Context, client Client) (any, error) {
		return nil, client.UpdateGeonames(ctx)
	})
	return err
}

func (c *MultiClient) Hosting(ctx context.Context, address string) (*entity.Hosting, error) {
	return getFromAny(ctx, c.Clients, func(ctx context.Context, client Client) (*entity.Hosting, error) {
		return client.Hosting(ctx, address)
	})
}
