package geonames

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
	dir string

	countries    *geonameEntityStorage[*entity.GeoNameCountry]
	subdivisions *geonameEntityStorage[*entity.GeoNameAdminSubdivision]
	cities       *geonameEntityStorage[*entity.GeoName]

	additionalInfoReadyC chan struct{}
}

func NewStorage(dir string) *GeoNameStorage {
	s := new(GeoNameStorage)
	s.dir = dir

	s.fill()

	return s
}

func (s *GeoNameStorage) fill() {
	s.countries = newGeonameEntityStorage(s.dir, func(parser geonames.Parser) ([]*entity.GeoNameCountry, error) {
		var countries []*entity.GeoNameCountry
		err := parser.GetCountries(func(c *models.Country) error {
			countries = append(countries, &entity.GeoNameCountry{Country: c})
			return nil
		})
		return countries, err
	})

	s.subdivisions = newGeonameEntityStorage(s.dir, func(parser geonames.Parser) ([]*entity.GeoNameAdminSubdivision, error) {
		var subdivisions []*entity.GeoNameAdminSubdivision
		err := parser.GetAdminDivisions(func(division *models.AdminDivision) error {
			subdivisions = append(subdivisions, &entity.GeoNameAdminSubdivision{AdminDivision: division})
			return nil
		})
		return subdivisions, err
	})

	s.cities = newGeonameEntityStorage(s.dir, func(parser geonames.Parser) ([]*entity.GeoName, error) {
		var cities []*entity.GeoName
		err := parser.GetGeonames(geonames.Cities500, func(c *models.Geoname) error {
			cities = append(cities, &entity.GeoName{Geoname: c})
			return nil
		})
		return cities, err
	})

	s.additionalInfoReadyC = make(chan struct{})
	go s.fillAdditionalFields()
}

func (s *GeoNameStorage) CheckUpdates(ctx context.Context) (entity.Updates, error) {
	return nil, nil
}

func (s *GeoNameStorage) Download(ctx context.Context) (entity.Updates, error) {
	s.fill()
	return entity.Updates{
		entity.SubjectGeonames: &entity.UpdateStatus{Error: ""},
	}, nil
}

func (r *GeoNameStorage) fillAdditionalFields() {
	defer close(r.additionalInfoReadyC)
	r.waitInit()

	countryCodeToContinent := make(map[string]*entity.GeoNameContinent)
	countryCodeToCountry := make(map[string]*entity.GeoNameCountry)
	subdivisionCodeToSubdivision := make(map[string]*entity.GeoNameAdminSubdivision)

	for _, country := range r.countries.collection {
		var continent *entity.GeoNameContinent
		for _, c := range r.Continents(context.Background()) {
			if c.Code() == country.Continent {
				continent = c
				break
			}
		}

		country.ContinentName = continent.GetContinentName()

		countryCodeToCountry[country.GetCountryCode()] = country
		countryCodeToContinent[country.GetCountryCode()] = continent
	}

	for _, subdiv := range r.subdivisions.collection {
		if country, ok := countryCodeToCountry[subdiv.GetCountryCode()]; ok {
			subdiv.CountryName = country.GetName()
		}

		if continent, ok := countryCodeToContinent[subdiv.GetCountryCode()]; ok {
			subdiv.ContinentCode = continent.GetContinentCode()
			subdiv.ContinentName = continent.GetName()
		}

		subdivisionCodeToSubdivision[subdiv.Code] = subdiv
	}

	for _, city := range r.cities.collection {
		if country, ok := countryCodeToCountry[city.GetCountryCode()]; ok {
			city.CountryCode = country.GetCountryCode()
			city.CountryName = country.GetCountryName()
		}
		if continent, ok := countryCodeToContinent[city.GetCountryCode()]; ok {
			city.ContinentCode = continent.GetContinentCode()
			city.ContinentName = continent.GetName()
		}
		if subdiv, ok := subdivisionCodeToSubdivision[city.CountryCode+"."+city.Admin1Code]; ok {
			city.SubdivisionName = subdiv.GetSubdivisionName()
		}
	}
}

func (r *GeoNameStorage) waitInit() {
	<-r.cities.readyC
	<-r.subdivisions.readyC
	<-r.countries.readyC
}

func (r *GeoNameStorage) WaitReady() {
	<-r.additionalInfoReadyC
}

func (r *GeoNameStorage) Continents(ctx context.Context) []*entity.GeoNameContinent {
	return GeoNameContinents()
}

func (r *GeoNameStorage) Countries(ctx context.Context, filter entity.GeoNameFilter) ([]*entity.GeoNameCountry, error) {
	select {
	case <-r.additionalInfoReadyC:
		return r.countries.GetEntities(ctx, filter)
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (r *GeoNameStorage) Subdivisions(ctx context.Context, filter entity.GeoNameFilter) ([]*entity.GeoNameAdminSubdivision, error) {
	select {
	case <-r.additionalInfoReadyC:
		return r.subdivisions.GetEntities(ctx, filter)
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (r *GeoNameStorage) Cities(ctx context.Context, filter entity.GeoNameFilter) ([]*entity.GeoName, error) {
	select {
	case <-r.additionalInfoReadyC:
		return r.cities.GetEntities(ctx, filter)
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
