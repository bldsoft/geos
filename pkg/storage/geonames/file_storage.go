package geonames

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/bldsoft/geos/pkg/entity"
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

	readyC chan struct{}
}

func newGeonameEntityStorage[T entity.GeoNameEntity](dir string, fillCollectionCallback func(parser geonames.Parser) ([]T, error)) *geonameEntityStorage[T] {
	s := &geonameEntityStorage[T]{
		index:                  &index[T]{},
		readyC:                 make(chan struct{}),
		fillCollectionCallback: fillCollectionCallback,
	}
	go s.init(dir)
	return s
}

func (s *geonameEntityStorage[T]) init(dir string) {
	defer close(s.readyC)
	if len(dir) == 0 {
		return
	}

	parser := geonames.Parser(func(filename string) (io.ReadCloser, error) {
		fullpath := filepath.Join(dir, filename)

		if file, err := os.OpenFile(fullpath, os.O_RDONLY, 0666); err == nil {
			return file, nil
		}

		reader, err := geonames.NewParser()(filename)
		if err != nil {
			return nil, nil
		}

		file, err := os.OpenFile(fullpath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			log.Logger.ErrorWithFields(log.Fields{"path": fullpath, "err": err}, "Geonames: failed to create file")
			return reader, nil
		}

		return utils.NewTeeReadCloser(reader, file), nil
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
	select {
	case <-s.readyC:
		if len(s.collection) == 0 {
			return nil, ErrGeoNameDisabled
		}
		filtered := s.index.GetFiltered(filter)
		if filter.Limit != 0 && len(filtered) > int(filter.Limit) {
			return filtered[:filter.Limit], nil
		}
		return filtered, nil
	default:
		return nil, ErrGeoNameNotReady
	}
}
