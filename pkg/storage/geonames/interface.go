package geonames

import (
	"context"

	"github.com/bldsoft/geos/pkg/entity"
)

type Storage interface {
	Continents(ctx context.Context) []*entity.GeoNameContinent
	Countries(ctx context.Context, filter entity.GeoNameFilter) ([]*entity.GeoNameCountry, error)
	Subdivisions(ctx context.Context, filter entity.GeoNameFilter) ([]*entity.GeoNameAdminSubdivision, error)
	Cities(ctx context.Context, filter entity.GeoNameFilter) ([]*entity.GeoName, error)
}
