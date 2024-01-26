package repository

import (
	"bytes"
	"context"
	"encoding/csv"
	"strconv"

	"github.com/bldsoft/geos/pkg/entity"
)

type GeoNameStorage interface {
	Continents(ctx context.Context) []*entity.GeoNameContinent
	Countries(ctx context.Context, filter entity.GeoNameFilter) ([]*entity.GeoNameCountry, error)
	Subdivisions(ctx context.Context, filter entity.GeoNameFilter) ([]*entity.GeoNameAdminSubdivision, error)
	Cities(ctx context.Context, filter entity.GeoNameFilter) ([]*entity.GeoName, error)
}

type GeoNameRepository struct {
	storage GeoNameStorage
}

func NewGeoNamesRepository(storage GeoNameStorage) *GeoNameRepository {
	return &GeoNameRepository{storage: storage}
}

func (r *GeoNameRepository) Continents(ctx context.Context) []*entity.GeoNameContinent {
	return r.storage.Continents(ctx)
}

func (r *GeoNameRepository) Countries(ctx context.Context, filter entity.GeoNameFilter) ([]*entity.GeoNameCountry, error) {
	return r.storage.Countries(ctx, filter)
}

func (r *GeoNameRepository) Subdivisions(ctx context.Context, filter entity.GeoNameFilter) ([]*entity.GeoNameAdminSubdivision, error) {
	return r.storage.Subdivisions(ctx, filter)
}

func (r *GeoNameRepository) Cities(ctx context.Context, filter entity.GeoNameFilter) ([]*entity.GeoName, error) {
	return r.storage.Cities(ctx, filter)
}

func (r *GeoNameRepository) Dump(ctx context.Context, format DumpFormat) ([]byte, error) {
	var buf bytes.Buffer
	csvWriter := csv.NewWriter(&buf)
	if format != DumpFormatCSV {
		if err := csvWriter.Write([]string{
			"geoname_id",
			"continent_code",
			"continent_name",
			"country_iso_code",
			"country_name",
			"subdivision_name",
			"city_name",
			"time_zone",
		}); err != nil {
			return nil, err
		}
	}

	cities, err := r.Cities(ctx, entity.GeoNameFilter{})
	if err != nil {
		return nil, err
	}
	for _, city := range cities {
		countries, err := r.Countries(ctx, entity.GeoNameFilter{CountryCodes: []string{city.CountryCode()}})
		if err != nil {
			return nil, err
		}
		country := countries[0]

		var continent *entity.GeoNameContinent
		for _, c := range r.Continents(ctx) {
			if c.Code() == country.Continent {
				continent = c
				break
			}
		}

		subdivisions, err := r.Subdivisions(ctx, entity.GeoNameFilter{CountryCodes: []string{city.CountryCode()}})
		if err != nil {
			return nil, err
		}

		var subdivisionName string
		for _, sub := range subdivisions {
			if sub.AdminCode() == city.Admin1Code {
				subdivisionName = sub.Name()
				break
			}
		}

		if err := csvWriter.Write([]string{
			strconv.Itoa(city.GeoNameID()),
			continent.Code(),
			continent.Name(),
			city.CountryCode(),
			country.Name(),
			subdivisionName,
			city.Name(),
			city.Timezone,
		}); err != nil {
			return nil, err
		}
	}

	if err := csvWriter.Write([]string{
		"9999999",
		"PRIVATE",
		"Private",
		"PRIVATE",
		"PRIVATE",
		"Private",
		"PRIVATE",
		"Africa/Harare",
	}); err != nil {
		return nil, err
	}

	csvWriter.Flush()
	return buf.Bytes(), nil
}
