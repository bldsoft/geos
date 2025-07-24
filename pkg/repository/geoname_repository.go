package repository

import (
	"bytes"
	"context"
	"encoding/csv"
	"strconv"
	"time"

	"github.com/bldsoft/geos/pkg/entity"
	"github.com/bldsoft/geos/pkg/storage/geonames"
	"github.com/bldsoft/geos/pkg/storage/source"
	"go.uber.org/atomic"
	"golang.org/x/sync/singleflight"
)

const GeonamesDBType = "geonames"

type StorageConfig struct {
	LocalDir         string
	PatchesRemoteURL string
	AutoUpdatePeriod time.Duration
}
type GeoNameRepository struct {
	cfg     StorageConfig
	storage geonames.Storage

	storageUpdater *updaterWithLastErr

	*baseUpdateRepository
	checkUpdatesSF  singleflight.Group
	lastUpdateError atomic.Pointer[string]
}

func NewGeoNamesRepository(config StorageConfig) *GeoNameRepository {
	origSource := source.NewGeoNamesSource(config.LocalDir)
	patchSource := source.NewPatchesSource(config.PatchesRemoteURL, config.LocalDir, GeonamesDBType)

	original := geonames.NewStorage(origSource)
	custom := geonames.NewCustomStorageFromTarGz(patchSource)

	var storage geonames.Storage = geonames.NewMultiStorage[geonames.Storage](original).Add(custom)
	res := &GeoNameRepository{
		cfg:            config,
		storage:        storage,
		storageUpdater: newUpdaterWithLastErr(storage.Update),
	}
	res.baseUpdateRepository = NewBaseUpdateRepository(
		"geonames.lock",
		config.AutoUpdatePeriod,
		res.storageUpdater.Update,
	)
	return res
}

func (r *GeoNameRepository) Run(ctx context.Context) error {
	return r.baseUpdateRepository.Run(ctx)
}

func (r *GeoNameRepository) StartUpdate(ctx context.Context) error {
	return r.baseUpdateRepository.StartUpdate(ctx)
}

func (r *GeoNameRepository) CheckUpdates(ctx context.Context) (entity.DBUpdate, error) {
	result, err, _ := r.checkUpdatesSF.Do("check_updates", func() (interface{}, error) {
		updates, err := r.storage.CheckUpdates(ctx)
		if err != nil {
			return entity.DBUpdate{}, err
		}
		return entity.NewDBUpdate(
			GeonamesDBType,
			updates,
			r.storageUpdater.InProgress(),
			r.storageUpdater.LastErr(),
		), nil
	})

	if err != nil {
		return entity.DBUpdate{}, err
	}
	return result.(entity.DBUpdate), nil
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

	if err = writeEntitiesToCSV(csvWriter, cities); err != nil {
		return nil, err
	}

	subdivs, err := r.Subdivisions(ctx, entity.GeoNameFilter{})
	if err != nil {
		return nil, err
	}

	if err = writeEntitiesToCSV(csvWriter, subdivs); err != nil {
		return nil, err
	}

	countries, err := r.Countries(ctx, entity.GeoNameFilter{})
	if err != nil {
		return nil, err
	}

	if err = writeEntitiesToCSV(csvWriter, countries); err != nil {
		return nil, err
	}

	continents := r.Continents(ctx)
	if err := writeEntitiesToCSV(csvWriter, continents); err != nil {
		return nil, err
	}

	if err := csvWriter.Write([]string{
		strconv.FormatUint(uint64(entity.PrivateCity.City.GeoNameID), 10),
		entity.PrivateCity.Continent.Code,
		entity.PrivateCity.Continent.Names["en"],
		entity.PrivateCity.Country.IsoCode,
		entity.PrivateCity.Country.Names["en"],
		entity.PrivateCity.Subdivisions[0].Names["en"],
		entity.PrivateCity.City.Names["en"],
		entity.PrivateCity.Location.TimeZone,
	}); err != nil {
		return nil, err
	}

	csvWriter.Flush()
	return buf.Bytes(), nil
}

func writeEntitiesToCSV[T entity.GeoNameEntity](w *csv.Writer, entities []T) error {
	for _, e := range entities {
		if err := w.Write([]string{
			strconv.Itoa(e.GetGeoNameID()),
			e.GetContinentCode(),
			e.GetContinentName(),
			e.GetCountryCode(),
			e.GetCountryName(),
			e.GetSubdivisionName(),
			e.GetCityName(),
			e.GetTimeZone(),
		}); err != nil {
			return err
		}
	}
	return nil
}
