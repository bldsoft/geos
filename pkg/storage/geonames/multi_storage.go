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
	limited := filter.Limit > 0
	var res []*entity.GeoNameCountry
	for _, s := range s.storages {
		countries, err := s.Countries(ctx, filter)
		if err != nil {
			return nil, err
		}
		res = append(res, countries...)
		if consumeLimit(&filter, len(countries), limited) {
			break
		}
	}
	return res, nil
}

func (s *MultiStorage) Subdivisions(ctx context.Context, filter entity.GeoNameFilter) ([]*entity.GeoNameAdminSubdivision, error) {
	limited := filter.Limit > 0
	var res []*entity.GeoNameAdminSubdivision
	for _, s := range s.storages {
		subdivisions, err := s.Subdivisions(ctx, filter)
		if err != nil {
			return nil, err
		}
		res = append(res, subdivisions...)
		if consumeLimit(&filter, len(subdivisions), limited) {
			break
		}
	}
	return res, nil
}

func (s *MultiStorage) Cities(ctx context.Context, filter entity.GeoNameFilter) ([]*entity.GeoName, error) {
	limited := filter.Limit > 0
	var res []*entity.GeoName
	for _, s := range s.storages {
		cities, err := s.Cities(ctx, filter)
		if err != nil {
			return nil, err
		}
		res = append(res, cities...)
		if consumeLimit(&filter, len(cities), limited) {
			break
		}
	}
	return res, nil
}

func consumeLimit(filter *entity.GeoNameFilter, n int, limited bool) (exhausted bool) {
	if !limited {
		return false
	}
	if n >= int(filter.Limit) {
		filter.Limit = 0

		return true
	}
	filter.Limit -= uint32(n)

	return false
}
