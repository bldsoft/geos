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
		{GeonameID: AF, Code: "AF", Name: "Africa"},
		{GeonameID: AS, Code: "AS", Name: "Asia"},
		{GeonameID: EU, Code: "EU", Name: "Europe"},
		{GeonameID: NA, Code: "NA", Name: "North America"},
		{GeonameID: OC, Code: "OC", Name: "Oceania"},
		{GeonameID: SA, Code: "SA", Name: "South America"},
		{GeonameID: AN, Code: "AN", Name: "Antarctica"},
	}
}

type GeoNameStorage struct {
	countries    *geonameEntityStorage[entity.GeoNameCountry]
	subdivisions *geonameEntityStorage[entity.AdminSubdivision]
	cities       *geonameEntityStorage[entity.Geoname]
}

func NewGeoNamesStorage(dir string) *GeoNameStorage {
	s := &GeoNameStorage{
		countries: newGeonameEntityStorage(dir, func(parser geonames.Parser) ([]*entity.GeoNameCountry, geoNameIndex, nameIndex, error) {
			countries, index := []*entity.GeoNameCountry{}, newGeoNameIndex()
			var nameIndex nameIndex
			err := parser.GetCountries(func(c *models.Country) error {
				countries = append(countries, (*entity.GeoNameCountry)(c))
				index.put(c.Iso2Code)
				nameIndex.names = append(nameIndex.names, c.Name)
				return nil
			})
			return countries, index, nameIndex, err
		}),
		subdivisions: newGeonameEntityStorage(dir, func(parser geonames.Parser) ([]*entity.AdminSubdivision, geoNameIndex, nameIndex, error) {
			subdivisions, index := []*entity.AdminSubdivision{}, newGeoNameIndex()
			var nameIndex nameIndex
			err := parser.GetAdminDivisions(func(division *models.AdminDivision) error {
				res := (*entity.AdminSubdivision)(division)
				subdivisions = append(subdivisions, res)
				index.put(res.CountryCode())
				nameIndex.names = append(nameIndex.names, division.Name)
				return nil
			})
			return subdivisions, index, nameIndex, err
		}),
		cities: newGeonameEntityStorage(dir, func(parser geonames.Parser) ([]*entity.Geoname, geoNameIndex, nameIndex, error) {
			cities, index := []*entity.Geoname{}, newGeoNameIndex()
			var nameIndex nameIndex
			err := parser.GetGeonames(geonames.Cities500, func(c *models.Geoname) error {
				cities = append(cities, (*entity.Geoname)(c))
				index.put(c.CountryCode)
				nameIndex.names = append(nameIndex.names, c.Name)
				return nil
			})
			return cities, index, nameIndex, err
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

func (r *GeoNameStorage) Subdivisions(ctx context.Context, filter entity.GeoNameFilter) ([]*entity.AdminSubdivision, error) {
	return r.subdivisions.GetEntities(ctx, filter)
}

func (r *GeoNameStorage) Cities(ctx context.Context, filter entity.GeoNameFilter) ([]*entity.Geoname, error) {
	return r.cities.GetEntities(ctx, filter)
}
