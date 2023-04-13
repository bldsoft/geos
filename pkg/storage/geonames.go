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

	countriesIndex    geoNameIndex
	subdivisionsIndex geoNameIndex
	cityIndex         geoNameIndex

	readyC chan struct{}
}

func NewGeoNamesStorage(dir string) *GeoNameStorage {
	s := &GeoNameStorage{readyC: make(chan struct{}),
		countriesIndex:    newGeoNameIndex(),
		subdivisionsIndex: newGeoNameIndex(),
		cityIndex:         newGeoNameIndex(),
	}
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
				s.countriesIndex.put(c.Iso2Code)
				return nil
			}); err != nil {
				s.countries = nil
				s.countriesIndex = newGeoNameIndex()
				continue
			}
			log.Logger.Info("GeoNames countries loaded")
		}

		if s.subdivisions == nil {
			if err := parser.GetAdminSubdivisions(func(sub *models.AdminSubdivision) error {
				res := (*entity.AdminSubdivision)(sub)
				s.subdivisions = append(s.subdivisions, res)
				s.subdivisionsIndex.put(res.CountryCode())
				return nil
			}); err != nil {
				s.subdivisions = nil
				s.subdivisionsIndex = newGeoNameIndex()
				continue
			}
			log.Logger.Info("GeoNames subdivisions loaded")
		}

		if s.cities == nil {
			if err := parser.GetGeonames(geonames.Cities500, func(c *models.Geoname) error {
				s.cities = append(s.cities, (*entity.Geoname)(c))
				s.cityIndex.put(c.CountryCode)
				return nil
			}); err != nil {
				s.cities = nil
				s.cityIndex = newGeoNameIndex()
				continue
			}
			log.Logger.Info("GeoNames cities loaded")
		}
		return
	}

}

func (r *GeoNameStorage) Countries(ctx context.Context, filter entity.GeoNameFilter) ([]*entity.GeoNameCountry, error) {
	select {
	case <-r.readyC:
		if len(r.countries) == 0 {
			return nil, ErrGeoNameDisabled
		}
		if len(filter.CountryCodes) == 0 {
			return r.countries, nil
		}

		res := make([]*entity.GeoNameCountry, 0, len(filter.CountryCodes))
		for _, code := range filter.CountryCodes {
			idx := r.countriesIndex.indexRange(code)
			res = append(res, r.countries[idx.begin:idx.end]...)
		}
		return res, nil
	default:
		return nil, ErrGeoNameNotReady
	}
}

func (r *GeoNameStorage) Subdivisions(ctx context.Context, filter entity.GeoNameFilter) ([]*entity.AdminSubdivision, error) {
	select {
	case <-r.readyC:
		if len(r.subdivisions) == 0 {
			return nil, ErrGeoNameDisabled
		}
		if len(filter.CountryCodes) == 0 {
			return r.subdivisions, nil
		}

		var res []*entity.AdminSubdivision
		for _, code := range filter.CountryCodes {
			idx := r.subdivisionsIndex.indexRange(code)
			res = append(res, r.subdivisions[idx.begin:idx.end]...)
		}
		return res, nil
	default:
		return nil, ErrGeoNameNotReady
	}
}

func (r *GeoNameStorage) Cities(ctx context.Context, filter entity.GeoNameFilter) ([]*entity.Geoname, error) {
	select {
	case <-r.readyC:
		if len(r.cities) == 0 {
			return nil, ErrGeoNameDisabled
		}

		if len(filter.CountryCodes) == 0 {
			return r.cities, nil
		}

		var res []*entity.Geoname
		for _, code := range filter.CountryCodes {
			idx := r.cityIndex.indexRange(code)
			res = append(res, r.cities[idx.begin:idx.end]...)
		}
		return res, nil
	default:
		return nil, ErrGeoNameNotReady
	}
}
