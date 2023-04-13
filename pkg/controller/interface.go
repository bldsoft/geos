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
	Countries(ctx context.Context) ([]*entity.GeoNameCountry, error)
	Subdivisions(ctx context.Context) ([]*entity.AdminSubdivision, error)
	Cities(ctx context.Context) ([]*entity.Geoname, error)
}
