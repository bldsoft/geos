package geonames

import (
	"context"
	"sync/atomic"

	"github.com/bldsoft/geos/pkg/entity"
	"github.com/bldsoft/geos/pkg/storage/source"
	"github.com/bldsoft/geos/pkg/storage/state"
	"github.com/bldsoft/gost/utils/errgroup"
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
	source *source.GeoNamesSource

	countries    atomic.Pointer[geonameEntityStorage[*entity.GeoNameCountry]]
	subdivisions atomic.Pointer[geonameEntityStorage[*entity.GeoNameAdminSubdivision]]
	cities       atomic.Pointer[geonameEntityStorage[*entity.GeoName]]
}

func NewStorage(source *source.GeoNamesSource, syncInit ...bool) *GeoNameStorage {
	s := &GeoNameStorage{
		source: source,
	}

	if len(syncInit) > 0 && syncInit[0] {
		s.fill()
	} else {
		go s.fill()
	}

	return s
}

func (s *GeoNameStorage) fill() {
	var eg errgroup.Group

	var countries *geonameEntityStorage[*entity.GeoNameCountry]
	eg.Go(func() error {
		countries = newGeonameEntityStorage(s.source.CountriesFile, func(parser geonames.Parser) ([]*entity.GeoNameCountry, error) {
			var countries []*entity.GeoNameCountry
			err := parser.GetCountries(func(c *models.Country) error {
				countries = append(countries, &entity.GeoNameCountry{Country: c})
				return nil
			})
			return countries, err
		})
		return nil
	})

	var subdivisions *geonameEntityStorage[*entity.GeoNameAdminSubdivision]
	eg.Go(func() error {
		subdivisions = newGeonameEntityStorage(s.source.AdminDivisionsFile, func(parser geonames.Parser) ([]*entity.GeoNameAdminSubdivision, error) {
			var subdivisions []*entity.GeoNameAdminSubdivision
			err := parser.GetAdminDivisions(func(division *models.AdminDivision) error {
				subdivisions = append(subdivisions, &entity.GeoNameAdminSubdivision{AdminDivision: division})
				return nil
			})
			return subdivisions, err
		})
		return nil
	})

	var cities *geonameEntityStorage[*entity.GeoName]
	eg.Go(func() error {
		cities = newGeonameEntityStorage(s.source.Cities500File, func(parser geonames.Parser) ([]*entity.GeoName, error) {
			var cities []*entity.GeoName
			err := parser.GetGeonames(geonames.Cities500, func(c *models.Geoname) error {
				cities = append(cities, &entity.GeoName{Geoname: c})
				return nil
			})
			return cities, err
		})
		return nil
	})

	_ = eg.Wait()
	s.fillAdditionalFields(countries, subdivisions, cities)

	s.countries.Store(countries)
	s.subdivisions.Store(subdivisions)
	s.cities.Store(cities)
}

func (r *GeoNameStorage) fillAdditionalFields(
	countries *geonameEntityStorage[*entity.GeoNameCountry],
	subdivisions *geonameEntityStorage[*entity.GeoNameAdminSubdivision],
	cities *geonameEntityStorage[*entity.GeoName],
) {
	countryCodeToContinent := make(map[string]*entity.GeoNameContinent)
	countryCodeToCountry := make(map[string]*entity.GeoNameCountry)
	subdivisionCodeToSubdivision := make(map[string]*entity.GeoNameAdminSubdivision)

	for _, country := range countries.collection {
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

	for _, subdiv := range subdivisions.collection {
		if country, ok := countryCodeToCountry[subdiv.GetCountryCode()]; ok {
			subdiv.CountryName = country.GetName()
		}

		if continent, ok := countryCodeToContinent[subdiv.GetCountryCode()]; ok {
			subdiv.ContinentCode = continent.GetContinentCode()
			subdiv.ContinentName = continent.GetName()
		}

		subdivisionCodeToSubdivision[subdiv.Code] = subdiv
	}

	for _, city := range cities.collection {
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

func (r *GeoNameStorage) Continents(ctx context.Context) []*entity.GeoNameContinent {
	return GeoNameContinents()
}

func (r *GeoNameStorage) Countries(ctx context.Context, filter entity.GeoNameFilter) ([]*entity.GeoNameCountry, error) {
	countries := r.countries.Load()
	if countries == nil {
		return nil, ErrGeoNameNotReady
	}
	return countries.GetEntities(ctx, filter)
}

func (r *GeoNameStorage) Subdivisions(ctx context.Context, filter entity.GeoNameFilter) ([]*entity.GeoNameAdminSubdivision, error) {
	subdivisions := r.subdivisions.Load()
	if subdivisions == nil {
		return nil, ErrGeoNameNotReady
	}
	return subdivisions.GetEntities(ctx, filter)
}

func (r *GeoNameStorage) Cities(ctx context.Context, filter entity.GeoNameFilter) ([]*entity.GeoName, error) {
	cities := r.cities.Load()
	if cities == nil {
		return nil, ErrGeoNameNotReady
	}
	return cities.GetEntities(ctx, filter)
}

func (r *GeoNameStorage) State() *state.GeosState {
	var timestampsSum int64
	for _, file := range []*source.UpdatableFile[source.ModTimeVersion]{
		r.source.CountriesFile,
		r.source.AdminDivisionsFile,
		r.source.Cities500File,
	} {
		if version, err := file.Version(context.Background()); err == nil {
			timestampsSum += version.Time().Unix()
		}
	}

	return &state.GeosState{
		GeonamesTimestamps: timestampsSum,
	}
}

func (s *GeoNameStorage) CheckUpdates(ctx context.Context) (entity.Update, error) {
	return s.source.CheckUpdates(ctx)
}

func (s *GeoNameStorage) Update(ctx context.Context, force bool) error {
	update, err := s.CheckUpdates(ctx)
	if err != nil {
		return err
	}
	if update.AvailableVersion == "" {
		return nil
	}

	if err := s.source.Update(ctx, force); err != nil {
		return err
	}

	s.fill()
	return nil
}
