package service

import (
	"context"

	"github.com/bldsoft/geos/pkg/entity"
)

type GeoNameRepository interface {
	Countries(ctx context.Context, filter entity.GeoNameFilter) ([]*entity.GeoNameCountry, error)
	Subdivisions(ctx context.Context, filter entity.GeoNameFilter) ([]*entity.AdminSubdivision, error)
	Cities(ctx context.Context, filter entity.GeoNameFilter) ([]*entity.Geoname, error)
}

type GeoNameService struct {
	GeoNameRepository
}

func NewGeoNameService(rep GeoNameRepository) *GeoNameService {
	return &GeoNameService{rep}
}

func (r *GeoNameService) Countries(ctx context.Context, filter entity.GeoNameFilter) ([]*entity.GeoNameCountry, error) {
	return r.GeoNameRepository.Countries(ctx, filter)
}

func (r *GeoNameService) Subdivisions(ctx context.Context, filter entity.GeoNameFilter) ([]*entity.AdminSubdivision, error) {
	return r.GeoNameRepository.Subdivisions(ctx, filter)
}

func (r *GeoNameService) Cities(ctx context.Context, filter entity.GeoNameFilter) ([]*entity.Geoname, error) {
	return r.GeoNameRepository.Cities(ctx, filter)
}
