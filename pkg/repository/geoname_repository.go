package repository

import (
	"context"
	"errors"

	"github.com/bldsoft/geos/pkg/entity"
)

var ErrNotReady = errors.New("not ready")

type GeoNameStorage interface {
	Countries(ctx context.Context) ([]*entity.GeoNameCountry, error)
	Subdivisions(ctx context.Context) ([]*entity.AdminSubdivision, error)
	Cities(ctx context.Context) ([]*entity.Geoname, error)
}

type GeoNameRepository struct {
	storage GeoNameStorage
}

func NewGeoNamesRepository(storage GeoNameStorage) *GeoNameRepository {
	return &GeoNameRepository{storage: storage}
}

func (r *GeoNameRepository) Countries(ctx context.Context) ([]*entity.GeoNameCountry, error) {
	return r.storage.Countries(ctx)
}

func (r *GeoNameRepository) Subdivisions(ctx context.Context) ([]*entity.AdminSubdivision, error) {
	return r.storage.Subdivisions(ctx)
}

func (r *GeoNameRepository) Cities(ctx context.Context) ([]*entity.Geoname, error) {
	return r.storage.Cities(ctx)
}
