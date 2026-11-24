package service

import (
	"context"

	"github.com/bldsoft/geos/pkg/entity"
)

const maxSearchLimit = 100

type GeoNameRepository interface {
	Continents(ctx context.Context) []*entity.GeoNameContinent
	Countries(ctx context.Context, filter entity.GeoNameFilter) ([]*entity.GeoNameCountry, error)
	Subdivisions(ctx context.Context, filter entity.GeoNameFilter) ([]*entity.GeoNameAdminSubdivision, error)
	Cities(ctx context.Context, filter entity.GeoNameFilter) ([]*entity.GeoName, error)
	Dump(ctx context.Context, format DumpFormat) ([]byte, error)

	StartUpdate(ctx context.Context) error
	CheckUpdates(ctx context.Context) (entity.DBUpdate[entity.PatchedGeoNamesVersion], error)
	CurrentVersion(ctx context.Context) (entity.PatchedGeoNamesVersion, error)
}

type GeoNameService struct {
	GeoNameRepository
}

func NewGeoNameService(rep GeoNameRepository) *GeoNameService {
	return &GeoNameService{rep}
}

func (s *GeoNameService) CheckUpdates(ctx context.Context) (entity.DBUpdate[entity.PatchedGeoNamesVersion], error) {
	return s.GeoNameRepository.CheckUpdates(ctx)
}

func (s *GeoNameService) CurrentVersion(ctx context.Context) (entity.PatchedGeoNamesVersion, error) {
	return s.GeoNameRepository.CurrentVersion(ctx)
}

func (s *GeoNameService) StartUpdate(ctx context.Context) error {
	return s.GeoNameRepository.StartUpdate(ctx)
}

func (s *GeoNameService) Continents(ctx context.Context) []*entity.GeoNameContinent {
	return s.GeoNameRepository.Continents(ctx)
}

func (s *GeoNameService) Countries(ctx context.Context, filter entity.GeoNameFilter) ([]*entity.GeoNameCountry, error) {
	return s.GeoNameRepository.Countries(ctx, filter)
}

func (s *GeoNameService) Subdivisions(ctx context.Context, filter entity.GeoNameFilter) ([]*entity.GeoNameAdminSubdivision, error) {
	return s.GeoNameRepository.Subdivisions(ctx, s.limited(filter))
}

func (s *GeoNameService) Cities(ctx context.Context, filter entity.GeoNameFilter) ([]*entity.GeoName, error) {
	return s.GeoNameRepository.Cities(ctx, s.limited(filter))
}

func (s *GeoNameService) limited(filter entity.GeoNameFilter) entity.GeoNameFilter {
	if len(filter.GeoNameIDs) == 0 && (filter.Limit == 0 || filter.Limit > maxSearchLimit) {
		filter.Limit = maxSearchLimit
	}
	return filter
}

func (s *GeoNameService) Dump(ctx context.Context, format DumpFormat) ([]byte, error) {
	return s.GeoNameRepository.Dump(ctx, format)
}
