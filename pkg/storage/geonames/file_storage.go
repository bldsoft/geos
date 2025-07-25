package geonames

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/bldsoft/geos/pkg/entity"
	"github.com/bldsoft/geos/pkg/storage/source"
	"github.com/bldsoft/geos/pkg/utils"
	"github.com/bldsoft/gost/log"
	"github.com/mkrou/geonames"
)

const initRetryInterval = time.Minute

var (
	ErrGeoNameNotReady = fmt.Errorf("geoname csv dump is %w", utils.ErrNotReady)
	ErrGeoNameDisabled = fmt.Errorf("geoname csv dump is %w", utils.ErrDisabled)
)

type geonameIndex[T entity.GeoNameEntity] interface {
	Init(collection []T)
	GetFiltered(filter entity.GeoNameFilter) []T
}
type geonameEntityStorage[T entity.GeoNameEntity] struct {
	collection []T
	index      geonameIndex[T]

	fillCollectionCallback func(parser geonames.Parser) ([]T, error)
}

func newGeonameEntityStorage[T entity.GeoNameEntity](file *source.TSUpdatableFile, fillCollectionCallback func(parser geonames.Parser) ([]T, error)) *geonameEntityStorage[T] {
	s := &geonameEntityStorage[T]{
		index:                  &index[T]{},
		fillCollectionCallback: fillCollectionCallback,
	}
	s.init(file)
	return s
}

func (s *geonameEntityStorage[T]) init(file *source.TSUpdatableFile) {
	parser := geonames.Parser(func(filename string) (io.ReadCloser, error) {
		return file.Reader(context.Background())
	})

	ticker := time.NewTicker(initRetryInterval)
	defer ticker.Stop()
	for ; true; <-ticker.C {
		var err error
		s.collection, err = s.fillCollectionCallback(parser)
		if err == nil {
			s.index.Init(s.collection)
			break
		}
		log.Logger.ErrorWithFields(log.Fields{"err": err}, "Failed to load GeoNames dump")
	}
}

func (s *geonameEntityStorage[T]) GetEntities(ctx context.Context, filter entity.GeoNameFilter) ([]T, error) {
	if len(s.collection) == 0 {
		return nil, ErrGeoNameDisabled
	}
	filtered := s.index.GetFiltered(filter)
	if filter.Limit != 0 && len(filtered) > int(filter.Limit) {
		return filtered[:filter.Limit], nil
	}
	return filtered, nil
}
