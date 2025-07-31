package geonames

import (
	"context"

	"github.com/bldsoft/geos/pkg/entity"
)

type MultiStorage struct {
	storages []Storage
}

func NewMultiStorage(storages ...Storage) *MultiStorage {
	return &MultiStorage{storages: storages}
}

func (s *MultiStorage) Add(storages ...Storage) *MultiStorage {
	s.storages = append(s.storages, storages...)
	return s
}

func (s *MultiStorage) Continents(ctx context.Context) []*entity.GeoNameContinent {
	var res []*entity.GeoNameContinent
	for _, s := range s.storages {
		res = append(res, s.Continents(ctx)...)
	}
	return res
}

func (s *MultiStorage) Countries(ctx context.Context, filter entity.GeoNameFilter) ([]*entity.GeoNameCountry, error) {
	var res []*entity.GeoNameCountry
	for _, s := range s.storages {
		countries, err := s.Countries(ctx, filter)
		if err != nil {
			return nil, err
		}
		filter.Limit -= uint32(len(countries))
		res = append(res, countries...)
	}
	return res, nil
}

func (s *MultiStorage) Subdivisions(ctx context.Context, filter entity.GeoNameFilter) ([]*entity.GeoNameAdminSubdivision, error) {
	var res []*entity.GeoNameAdminSubdivision
	for _, s := range s.storages {
		subdivisions, err := s.Subdivisions(ctx, filter)
		if err != nil {
			return nil, err
		}
		filter.Limit -= uint32(len(subdivisions))
		res = append(res, subdivisions...)
	}
	return res, nil
}

func (s *MultiStorage) Cities(ctx context.Context, filter entity.GeoNameFilter) ([]*entity.GeoName, error) {
	var res []*entity.GeoName
	for _, s := range s.storages {
		cities, err := s.Cities(ctx, filter)
		if err != nil {
			return nil, err
		}
		filter.Limit -= uint32(len(cities))
		res = append(res, cities...)
	}
	return res, nil
}
