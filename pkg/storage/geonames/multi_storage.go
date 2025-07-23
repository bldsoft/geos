package geonames

import (
	"context"
	"errors"
	"maps"

	"github.com/bldsoft/geos/pkg/entity"
	"github.com/bldsoft/geos/pkg/storage/state"
)

type MultiStorage[T Storage] struct {
	storages []T
}

func NewMultiStorage[T Storage](storages ...T) *MultiStorage[T] {
	return &MultiStorage[T]{storages: storages}
}

func (s *MultiStorage[T]) CheckUpdates(ctx context.Context) (entity.Updates, error) {
	multiUpdates := entity.Updates{}
	var multiErr error
	for _, storage := range s.storages {
		updates, err := storage.CheckUpdates(ctx)
		if err != nil || updates == nil {
			multiErr = errors.Join(multiErr, err)
			continue
		}

		maps.Copy(multiUpdates, updates)
	}

	return multiUpdates, multiErr
}

func (s *MultiStorage[T]) Download(ctx context.Context) (entity.Updates, error) {
	multiUpdate := entity.Updates{}
	var multiErr error
	for _, storage := range s.storages {
		updates, err := storage.Download(ctx)
		if err != nil || updates == nil {
			multiErr = errors.Join(multiErr, err)
			continue
		}

		maps.Copy(multiUpdate, updates)
	}

	return multiUpdate, multiErr
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

func (s *MultiStorage[T]) State() *state.GeosState {
	result := &state.GeosState{}
	for _, storage := range s.storages {
		storageState := storage.State()
		result.Add(storageState)
	}
	return result
}

func (s *MultiStorage[T]) LastUpdateInterrupted(ctx context.Context) (bool, error) {
	for _, storage := range s.storages {
		interrupted, err := storage.LastUpdateInterrupted(ctx)
		if err != nil {
			return false, err
		}
		if interrupted {
			return true, nil
		}
	}
	return false, nil
}
