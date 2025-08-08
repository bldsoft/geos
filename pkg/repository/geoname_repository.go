package repository

import (
	"bytes"
	"context"
	"encoding/csv"
	"net/url"
	"path/filepath"
	"strconv"
	"time"

	"github.com/bldsoft/geos/pkg/entity"
	"github.com/bldsoft/geos/pkg/storage/geonames"
	"github.com/bldsoft/geos/pkg/storage/source"
	"github.com/bldsoft/gost/log"
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
	storage *geonames.PatchedStorage

	storageUpdater *updaterWithLastErr

	*baseUpdateRepository
	checkUpdatesSF singleflight.Group
}

func NewGeoNamesRepository(config StorageConfig) *GeoNameRepository {
	logger := log.Logger.WithFields(log.Fields{"db": "geonames"})
	ctx := context.WithValue(context.Background(), log.LoggerCtxKey, logger)

	origSource := source.NewGeoNamesSource(config.LocalDir)
	original := geonames.NewStorage(ctx, origSource)
	storage := geonames.NewPatchedStorage(original)

	if config.PatchesRemoteURL != "" {
		patchesURL, err := url.Parse(config.PatchesRemoteURL)
		if err != nil {
			log.FromContext(ctx).Fatalf("Failed to parse patches remote url: %s", err)
		}
		patchSource := source.NewTSUpdatableFile(
			filepath.Join(filepath.Dir(config.LocalDir), GeonamesDBType+"_patch"+filepath.Ext(patchesURL.Path)),
			config.PatchesRemoteURL,
		)
		custom := geonames.NewCustomStorage(ctx, patchSource)
		storage = storage.Add(custom)
	}

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

func (r *GeoNameRepository) CheckUpdates(ctx context.Context) (entity.DBUpdate[entity.PatchedGeoNamesVersion], error) {
	result, err, _ := r.checkUpdatesSF.Do("check_updates", func() (interface{}, error) {
		updates, err := r.storage.CheckUpdates(ctx)
		if err != nil {
			return entity.DBUpdate[entity.PatchedGeoNamesVersion]{}, err
		}
		return entity.NewDBUpdate(
			updates,
			r.IsInProgress(),
			r.LastErr(),
		), nil
	})
	return result.(entity.DBUpdate[entity.PatchedGeoNamesVersion]), err
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
