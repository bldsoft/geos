package service

import (
	"context"

	"github.com/bldsoft/geos/pkg/entity"
)

type GeoNameRepository interface {
	Continents(ctx context.Context) []*entity.GeoNameContinent
	Countries(ctx context.Context, filter entity.GeoNameFilter) ([]*entity.GeoNameCountry, error)
	Subdivisions(ctx context.Context, filter entity.GeoNameFilter) ([]*entity.GeoNameAdminSubdivision, error)
	Cities(ctx context.Context, filter entity.GeoNameFilter) ([]*entity.GeoName, error)
	Dump(ctx context.Context, format DumpFormat) ([]byte, error)
	Download(ctx context.Context, update ...bool) (entity.Updates, error)
	CheckUpdates(ctx context.Context) (entity.Updates, error)
}

type GeoNameService struct {
	GeoNameRepository
}

func NewGeoNameService(rep GeoNameRepository) *GeoNameService {
	return &GeoNameService{rep}
}

func (s *GeoNameService) Download(ctx context.Context, update ...bool) (entity.Updates, error) {
	return s.GeoNameRepository.Download(ctx, update...)
}

func (s *GeoNameService) CheckUpdates(ctx context.Context) (entity.Updates, error) {
	return s.GeoNameRepository.CheckUpdates(ctx)
}

func (s *GeoNameService) Continents(ctx context.Context) []*entity.GeoNameContinent {
	return s.GeoNameRepository.Continents(ctx)
}

func (s *GeoNameService) Countries(ctx context.Context, filter entity.GeoNameFilter) ([]*entity.GeoNameCountry, error) {
	return s.GeoNameRepository.Countries(ctx, filter)
}

func (s *GeoNameService) Subdivisions(ctx context.Context, filter entity.GeoNameFilter) ([]*entity.GeoNameAdminSubdivision, error) {
	return s.GeoNameRepository.Subdivisions(ctx, filter)
}

func (s *GeoNameService) Cities(ctx context.Context, filter entity.GeoNameFilter) ([]*entity.GeoName, error) {
	return s.GeoNameRepository.Cities(ctx, filter)
}

func (s *GeoNameService) Dump(ctx context.Context, format DumpFormat) ([]byte, error) {
	return s.GeoNameRepository.Dump(ctx, format)
}
