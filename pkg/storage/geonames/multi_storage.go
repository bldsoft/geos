package geonames

import (
	"context"
	"errors"

	"github.com/bldsoft/geos/pkg/entity"
)

type MultiStorage[T Storage] struct {
	storages []T
}

func NewMultiStorage[T Storage](storages ...T) *MultiStorage[T] {
	return &MultiStorage[T]{storages: storages}
}

func (s *MultiStorage[T]) CheckUpdates(ctx context.Context) (bool, error) {
	var multiErr error
	for _, storage := range s.storages {
		updated, err := storage.CheckUpdates(ctx)
		if err != nil {
			multiErr = errors.Join(multiErr, err)
			continue
		}
		if updated {
			return true, nil
		}
	}

	return false, multiErr
}

func (s *MultiStorage[T]) Download(ctx context.Context, update ...bool) error {
	var multiErr error
	for _, storage := range s.storages {
		err := storage.Download(ctx, update...)
		multiErr = errors.Join(multiErr, err)
	}

	return multiErr
}

func (s *MultiStorage[T]) DirPath() string {
	// TODO: kinda not used yet... and kinda no proper way to retrieve it yet...
	// and kinda looks like it should not be there at all...
	// so how its just to satisfy the interface...
	return ""
}

func (s *MultiStorage[T]) Add(storages ...T) *MultiStorage[T] {
	s.storages = append(s.storages, storages...)
	return s
}

func (s *MultiStorage[T]) Continents(ctx context.Context) []*entity.GeoNameContinent {
	var res []*entity.GeoNameContinent
	for _, s := range s.storages {
		res = append(res, s.Continents(ctx)...)
	}
	return res
}

func (s *MultiStorage[T]) Countries(ctx context.Context, filter entity.GeoNameFilter) ([]*entity.GeoNameCountry, error) {
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

func (s *MultiStorage[T]) Subdivisions(ctx context.Context, filter entity.GeoNameFilter) ([]*entity.GeoNameAdminSubdivision, error) {
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

func (s *MultiStorage[T]) Cities(ctx context.Context, filter entity.GeoNameFilter) ([]*entity.GeoName, error) {
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
