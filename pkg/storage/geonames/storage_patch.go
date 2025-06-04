package geonames

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/bldsoft/geos/pkg/entity"
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

func NewStoragePatchesFromDir(dir, customStoragesPrefix string) []*StoragePatch {
	var customStorages []*StoragePatch
	err := filepath.WalkDir(dir, func(s string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if !d.Type().IsRegular() || !strings.HasPrefix(d.Name(), customStoragesPrefix) {
			return nil
		}

		path := filepath.Join(dir, d.Name())
		custom, err := NewStoragePatchFromFile(path)
		if err != nil {
			if errors.Is(err, utils.ErrUnknownFormat) {
				return nil
			}
			return err
		}
		customStorages = append(customStorages, custom)

		return nil
	})
	if err != nil {
		log.WarnWithFields(log.Fields{"err": err}, "failed to read custom geonames storage")
	}
	return customStorages
}

func NewStoragePatchFromFile(path string) (*StoragePatch, error) {
	if filepath.Ext(path) != ".json" {
		return nil, utils.ErrUnknownFormat
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var records []CustomGeonamesRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return nil, err
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

func (s *StoragePatch) CheckUpdates(_ context.Context) (bool, error) {
	return false, nil
}

func (s *StoragePatch) Download(ctx context.Context, update ...bool) error {
	return nil
}
