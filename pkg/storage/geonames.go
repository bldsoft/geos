package storage

import (
	"context"

	"github.com/bldsoft/geos/pkg/entity"
	"github.com/mkrou/geonames"
	"github.com/mkrou/geonames/models"
)

type GeoNameStorage struct {
	countries    *geonameEntityStorage[entity.GeoNameCountry]
	subdivisions *geonameEntityStorage[entity.AdminSubdivision]
	cities       *geonameEntityStorage[entity.Geoname]
}

func NewGeoNamesStorage(dir string) *GeoNameStorage {
	s := &GeoNameStorage{
		countries: newGeonameEntityStorage(dir, func(parser geonames.Parser) ([]*entity.GeoNameCountry, geoNameIndex, error) {
			countries, index := []*entity.GeoNameCountry{}, newGeoNameIndex()
			err := parser.GetCountries(func(c *models.Country) error {
				countries = append(countries, (*entity.GeoNameCountry)(c))
				index.put(c.Iso2Code)
				return nil
			})
			return countries, index, err
		}),
		subdivisions: newGeonameEntityStorage(dir, func(parser geonames.Parser) ([]*entity.AdminSubdivision, geoNameIndex, error) {
			subdivisions, index := []*entity.AdminSubdivision{}, newGeoNameIndex()
			err := parser.GetAdminSubdivisions(func(sub *models.AdminSubdivision) error {
				res := (*entity.AdminSubdivision)(sub)
				subdivisions = append(subdivisions, res)
				index.put(res.CountryCode())
				return nil
			})
			return subdivisions, index, err
		}),
		cities: newGeonameEntityStorage(dir, func(parser geonames.Parser) ([]*entity.Geoname, geoNameIndex, error) {
			cities, index := []*entity.Geoname{}, newGeoNameIndex()
			err := parser.GetGeonames(geonames.Cities500, func(c *models.Geoname) error {
				cities = append(cities, (*entity.Geoname)(c))
				index.put(c.CountryCode)
				return nil
			})
			return cities, index, err
		}),
	}
	return s
}

func (r *GeoNameStorage) Countries(ctx context.Context, filter entity.GeoNameFilter) ([]*entity.GeoNameCountry, error) {
	return r.countries.GetEntities(ctx, filter)
}

func (r *GeoNameStorage) Subdivisions(ctx context.Context, filter entity.GeoNameFilter) ([]*entity.AdminSubdivision, error) {
	return r.subdivisions.GetEntities(ctx, filter)
}

func (r *GeoNameStorage) Cities(ctx context.Context, filter entity.GeoNameFilter) ([]*entity.Geoname, error) {
	return r.cities.GetEntities(ctx, filter)
}
