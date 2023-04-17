package controller

import (
	"context"

	"github.com/bldsoft/geos/pkg/entity"
)

type GeoIpService interface {
	Country(ctx context.Context, address string) (*entity.Country, error)
	City(ctx context.Context, address string) (*entity.City, error)
	CityLite(ctx context.Context, address string, lang string) (*entity.CityLite, error)
}

type GeoNameService interface {
	Continents(ctx context.Context) []*entity.GeoNameContinent
	Countries(ctx context.Context, filter entity.GeoNameFilter) ([]*entity.GeoNameCountry, error)
	Subdivisions(ctx context.Context, filter entity.GeoNameFilter) ([]*entity.AdminSubdivision, error)
	Cities(ctx context.Context, filter entity.GeoNameFilter) ([]*entity.Geoname, error)
}
