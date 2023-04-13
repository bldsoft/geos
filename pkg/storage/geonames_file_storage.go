package storage

import (
	"context"
	"errors"
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
	ErrGeoNameNotReady = errors.New("geoname not ready")
	ErrGeoNameDisabled = errors.New("geoname is disabled")
)

type geonameEntityStorage[T any] struct {
	collection []*T
	index      geoNameIndex

	fillCollectionCallback func(parser geonames.Parser) ([]*T, geoNameIndex, error)

	readyC chan struct{}
}

func newGeonameEntityStorage[T any](dir string, fillCollectionCallback func(parser geonames.Parser) ([]*T, geoNameIndex, error)) *geonameEntityStorage[T] {
	s := &geonameEntityStorage[T]{readyC: make(chan struct{}), fillCollectionCallback: fillCollectionCallback}
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
		s.collection, s.index, err = s.fillCollectionCallback(parser)
		if err == nil {
			break
		}
		log.Logger.ErrorWithFields(log.Fields{"err": err}, "Failed to load GeoNames dump")
	}
}

func (s *geonameEntityStorage[T]) GetEntities(ctx context.Context, filter entity.GeoNameFilter) ([]*T, error) {
	select {
	case <-s.readyC:
		if len(s.collection) == 0 {
			return nil, ErrGeoNameDisabled
		}
		if len(filter.CountryCodes) == 0 {
			return s.collection, nil
		}

		res := make([]*T, 0, len(filter.CountryCodes))
		for _, code := range filter.CountryCodes {
			idx := s.index.indexRange(code)
			res = append(res, s.collection[idx.begin:idx.end]...)
		}
		return res, nil
	default:
		return nil, ErrGeoNameNotReady
	}
}
