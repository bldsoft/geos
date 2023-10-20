package client

import (
	"context"

	"github.com/bldsoft/geos/pkg/entity"
)

type GeoIPClient interface {
	Country(ctx context.Context, address string) (*entity.Country, error)
	City(ctx context.Context, address string, includeISP bool) (*entity.City, error)
	CityLite(ctx context.Context, address, lang string) (*entity.CityLite, error)
}

type GeoNameClient interface {
	GeoNameContinents(ctx context.Context) []*entity.GeoNameContinent
	GeoNameCountries(ctx context.Context, filter entity.GeoNameFilter) ([]*entity.GeoNameCountry, error)
	GeoNameSubdivisions(ctx context.Context, filter entity.GeoNameFilter) ([]*entity.GeoNameAdminSubdivision, error)
	GeoNameCities(ctx context.Context, filter entity.GeoNameFilter) ([]*entity.GeoName, error)
}

type Client interface {
	GeoIPClient
	GeoNameClient
}
