package storage

import (
	"context"

	"github.com/bldsoft/geos/pkg/entity"
	"github.com/mkrou/geonames"
	"github.com/mkrou/geonames/models"
)

type Continent int

const (
	AF = 6255146
	AS = 6255147
	EU = 6255148
	NA = 6255149
	OC = 6255151
	SA = 6255150
	AN = 6255152
)

func GeoNameContinents() []*entity.GeoNameContinent {
	return []*entity.GeoNameContinent{
		entity.NewGeoNameContinent(AF, "AF", "Africa"),
		entity.NewGeoNameContinent(AS, "AS", "Asia"),
		entity.NewGeoNameContinent(EU, "EU", "Europe"),
		entity.NewGeoNameContinent(NA, "NA", "North America"),
		entity.NewGeoNameContinent(OC, "OC", "Oceania"),
		entity.NewGeoNameContinent(SA, "SA", "South America"),
		entity.NewGeoNameContinent(AN, "AN", "Antarctica"),
	}
}

type GeoNameStorage struct {
	countries    *geonameEntityStorage[*entity.GeoNameCountry]
	subdivisions *geonameEntityStorage[*entity.GeoNameAdminSubdivision]
	cities       *geonameEntityStorage[*entity.GeoName]
}

func NewGeoNamesStorage(dir string) *GeoNameStorage {
	s := &GeoNameStorage{
		countries: newGeonameEntityStorage(dir, func(parser geonames.Parser) ([]*entity.GeoNameCountry, error) {
			var countries []*entity.GeoNameCountry
			err := parser.GetCountries(func(c *models.Country) error {
				countries = append(countries, &entity.GeoNameCountry{Country: c})
				return nil
			})
			return countries, err
		}),
		subdivisions: newGeonameEntityStorage(dir, func(parser geonames.Parser) ([]*entity.GeoNameAdminSubdivision, error) {
			var subdivisions []*entity.GeoNameAdminSubdivision
			err := parser.GetAdminDivisions(func(division *models.AdminDivision) error {
				subdivisions = append(subdivisions, &entity.GeoNameAdminSubdivision{AdminDivision: division})
				return nil
			})
			return subdivisions, err
		}),
		cities: newGeonameEntityStorage(dir, func(parser geonames.Parser) ([]*entity.GeoName, error) {
			var cities []*entity.GeoName
			err := parser.GetGeonames(geonames.Cities500, func(c *models.Geoname) error {
				cities = append(cities, &entity.GeoName{Geoname: c})
				return nil
			})
			return cities, err
		}),
	}
	return s
}

func (r *GeoNameStorage) WaitReady() {
	<-r.cities.readyC
	<-r.subdivisions.readyC
	<-r.countries.readyC
}

func (r *GeoNameStorage) Continents(ctx context.Context) []*entity.GeoNameContinent {
	return GeoNameContinents()
}

func (r *GeoNameStorage) Countries(ctx context.Context, filter entity.GeoNameFilter) ([]*entity.GeoNameCountry, error) {
	return r.countries.GetEntities(ctx, filter)
}

func (r *GeoNameStorage) Subdivisions(ctx context.Context, filter entity.GeoNameFilter) ([]*entity.GeoNameAdminSubdivision, error) {
	return r.subdivisions.GetEntities(ctx, filter)
}

func (r *GeoNameStorage) Cities(ctx context.Context, filter entity.GeoNameFilter) ([]*entity.GeoName, error) {
	return r.cities.GetEntities(ctx, filter)
}
