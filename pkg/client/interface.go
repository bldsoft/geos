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

type ManagementClient interface {
	CheckGeoIPCityUpdates(ctx context.Context) (entity.DBUpdate[entity.PatchedMMDBVersion], error)
	CheckGeoIPISPUpdates(ctx context.Context) (entity.DBUpdate[entity.PatchedMMDBVersion], error)
	CheckGeonamesUpdates(ctx context.Context) (entity.DBUpdate[entity.PatchedGeoNamesVersion], error)
	UpdateGeoIPCity(ctx context.Context) error
	UpdateGeoIPISP(ctx context.Context) error
	UpdateGeonames(ctx context.Context) error
}

type Client interface {
	GeoIPClient
	GeoNameClient
	ManagementClient
}
