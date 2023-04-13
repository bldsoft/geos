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
	"github.com/mkrou/geonames/models"
)

var ErrGeoNameNotReady = errors.New("geoname not ready")
var ErrGeoNameDisabled = errors.New("geoname is disabled")

type GeoNameStorage struct {
	countries    []*entity.GeoNameCountry
	subdivisions []*entity.AdminSubdivision
	cities       []*entity.Geoname
	readyC       chan struct{}
}

func NewGeoNamesStorage(dir string) *GeoNameStorage {
	s := &GeoNameStorage{readyC: make(chan struct{})}
	go s.init(dir)
	return s
}

func (s *GeoNameStorage) init(dir string) {
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

	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for ; true; <-ticker.C {
		if s.countries == nil {
			if err := parser.GetCountries(func(c *models.Country) error {
				s.countries = append(s.countries, (*entity.GeoNameCountry)(c))
				return nil
			}); err != nil {
				s.countries = nil
				continue
			}
			log.Logger.Info("GeoNames countries loaded")
		}

		if s.subdivisions == nil {
			if err := parser.GetAdminSubdivisions(func(sub *models.AdminSubdivision) error {
				s.subdivisions = append(s.subdivisions, (*entity.AdminSubdivision)(sub))
				return nil
			}); err != nil {
				s.subdivisions = nil
				continue
			}
			log.Logger.Info("GeoNames subdivisions loaded")
		}

		if s.cities == nil {
			if err := parser.GetGeonames(geonames.Cities500, func(c *models.Geoname) error {
				s.cities = append(s.cities, (*entity.Geoname)(c))
				return nil
			}); err != nil {
				s.cities = nil
				continue
			}
			log.Logger.Info("GeoNames cities loaded")
		}
		return
	}

}

func (r *GeoNameStorage) Countries(ctx context.Context) ([]*entity.GeoNameCountry, error) {
	select {
	case <-r.readyC:
		if len(r.countries) == 0 {
			return nil, ErrGeoNameDisabled
		}
		return r.countries, nil
	default:
		return nil, ErrGeoNameNotReady
	}
}

func (r *GeoNameStorage) Subdivisions(ctx context.Context) ([]*entity.AdminSubdivision, error) {
	select {
	case <-r.readyC:
		if len(r.subdivisions) == 0 {
			return nil, ErrGeoNameDisabled
		}
		return r.subdivisions, nil
	default:
		return nil, ErrGeoNameNotReady
	}
}

func (r *GeoNameStorage) Cities(ctx context.Context) ([]*entity.Geoname, error) {
	select {
	case <-r.readyC:
		if len(r.cities) == 0 {
			return nil, ErrGeoNameDisabled
		}
		return r.cities, nil
	default:
		return nil, ErrGeoNameNotReady
	}
}
