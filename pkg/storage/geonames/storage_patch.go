package geonames

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"

	"github.com/bldsoft/geos/pkg/entity"
	"github.com/bldsoft/geos/pkg/storage/source"
	"github.com/bldsoft/geos/pkg/utils"
	"github.com/bldsoft/gost/log"
	"github.com/mkrou/geonames/models"
)

type CustomGeonamesEntity struct {
	GeoNameID int    `json:"geoNameID"`
	Name      string `json:"name"`
}

type CustomGeonamesRecord struct {
	City        CustomGeonamesEntity `json:"city"`
	Subdivision CustomGeonamesEntity `json:"subdivision"`
	Country     CustomGeonamesEntity `json:"country"`
	Continent   CustomGeonamesEntity `json:"continent"`
}

func NewCustomGeonamesRecord(geoName string, geoNameID int) CustomGeonamesRecord {
	return CustomGeonamesRecord{
		City: CustomGeonamesEntity{
			GeoNameID: geoNameID,
			Name:      geoName,
		},
		Subdivision: CustomGeonamesEntity{
			GeoNameID: geoNameID,
			Name:      geoName,
		},
		Country: CustomGeonamesEntity{
			GeoNameID: geoNameID,
			Name:      geoName,
		},
		Continent: CustomGeonamesEntity{
			GeoNameID: geoNameID,
			Name:      geoName,
		},
	}
}

func (rec CustomGeonamesRecord) ContinentEntity() *entity.GeoNameContinent {
	return entity.NewGeoNameContinent(rec.Continent.GeoNameID, rec.Continent.Name, rec.Continent.Name)
}

func (rec CustomGeonamesRecord) CountryEntity() *entity.GeoNameCountry {
	return &entity.GeoNameCountry{
		Country: &models.Country{
			GeonameID: rec.Country.GeoNameID,
			Name:      rec.Country.Name,
			Iso2Code:  rec.Country.Name,
			Continent: rec.Continent.Name,
		},
	}
}

func (rec CustomGeonamesRecord) SubdivisionEntity() *entity.GeoNameAdminSubdivision {
	return &entity.GeoNameAdminSubdivision{
		AdminDivision: &models.AdminDivision{
			GeonameId: rec.Subdivision.GeoNameID,
			Name:      rec.Subdivision.Name,
			Code:      rec.Country.Name + "." + rec.Subdivision.Name,
		},
	}
}

func (rec CustomGeonamesRecord) CityEntity() *entity.GeoName {
	return &entity.GeoName{
		Geoname: &models.Geoname{
			Id:          rec.City.GeoNameID,
			Name:        rec.City.Name,
			CountryCode: rec.Country.Name,
			Admin1Code:  rec.Subdivision.Name,
		},
	}
}

type StoragePatch struct {
	continents   []*entity.GeoNameContinent
	countries    []*entity.GeoNameCountry
	subdivisions []*entity.GeoNameAdminSubdivision
	cities       []*entity.GeoName
}

func NewStoragePatchesFromTarGz(source *source.TSUpdatableFile) ([]Storage, error) {
	ctx := context.Background()

	var customStorages []Storage
	r, err := source.Reader(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to open custom geonames storage archive: %w", err)
	}
	defer r.Close()

	content, err := utils.UnpackTarGz(r)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack custom geonames storage archive: %w", err)
	}

	for filename, data := range content {
		if filepath.Ext(filename) != ".json" {
			log.WarnWithFields(log.Fields{"name": filename}, "skipping non-json file in custom geonames storage archive")
			continue
		}

		patch, err := newStoragePatchFromJSON(data)
		if err != nil {
			log.ErrorWithFields(log.Fields{"err": err, "name": filename}, "failed to unmarshal custom geonames record")
			continue
		}
		customStorages = append(customStorages, patch)
	}

	return customStorages, nil
}

func NewStoragePatchFromJSON(source *source.TSUpdatableFile) (*StoragePatch, error) {
	ctx := context.Background()

	r, err := source.Reader(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to open custom geonames storage archive: %w", err)
	}
	defer r.Close()

	content, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read custom geonames storage archive: %w", err)
	}

	return newStoragePatchFromJSON(content)
}

func newStoragePatchFromJSON(data []byte) (*StoragePatch, error) {
	var records []CustomGeonamesRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return nil, fmt.Errorf("failed to unmarshal custom geonames record: %w", err)
	}
	return NewStoragePatch(records), nil
}

func NewStoragePatch(records []CustomGeonamesRecord) *StoragePatch {
	res := &StoragePatch{}
	for _, rec := range records {
		res.continents = append(res.continents, rec.ContinentEntity())
		res.subdivisions = append(res.subdivisions, rec.SubdivisionEntity())
		res.countries = append(res.countries, rec.CountryEntity())
		res.cities = append(res.cities, rec.CityEntity())
	}
	return res
}

func (s *StoragePatch) Continents(_ context.Context) []*entity.GeoNameContinent {
	return s.continents
}

func customFilter[T entity.GeoNameEntity](items []T, filter entity.GeoNameFilter) []T {
	var res []T
	for _, item := range items {
		if filter.Match(item) {
			res = append(res, item)
		}
	}
	return res
}

func (s *StoragePatch) Countries(_ context.Context, filter entity.GeoNameFilter) ([]*entity.GeoNameCountry, error) {
	return customFilter(s.countries, filter), nil
}

func (s *StoragePatch) Subdivisions(_ context.Context, filter entity.GeoNameFilter) ([]*entity.GeoNameAdminSubdivision, error) {
	return customFilter(s.subdivisions, filter), nil
}

func (s *StoragePatch) Cities(_ context.Context, filter entity.GeoNameFilter) ([]*entity.GeoName, error) {
	return customFilter(s.cities, filter), nil
}
