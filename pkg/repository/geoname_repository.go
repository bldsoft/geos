package repository

import (
	"bytes"
	"context"
	"encoding/csv"
	"strconv"
	"sync"
	"time"

	"github.com/bldsoft/geos/pkg/entity"
	"github.com/bldsoft/geos/pkg/storage/geonames"
	"github.com/bldsoft/geos/pkg/storage/source"
	"github.com/bldsoft/geos/pkg/storage/state"
	"github.com/bldsoft/geos/pkg/utils"
	"github.com/bldsoft/gost/log"
	"golang.org/x/sync/singleflight"
)

type GeoNameRepository struct {
	storage        geonames.Storage
	updateM        sync.Mutex
	checkUpdatesSF singleflight.Group
}

type StorageConfig struct {
	DirPath          string
	Source           *source.GeoNamesSource
	PatchesSource    *source.PatchesSource
	AutoUpdatePeriod int
}

func NewGeoNamesRepository(config *StorageConfig) *GeoNameRepository {
	rep := &GeoNameRepository{}

	ctx := context.Background()

	performSourceCheck(ctx, config.Source, "geonames", config.AutoUpdatePeriod)
	performSourceCheck(ctx, config.PatchesSource, "geonames patches", config.AutoUpdatePeriod)

	original := geonames.NewStorage(config.DirPath)
	original.SetSource(config.Source)

	custom := geonames.NewCustomStorageFromTarGz(config.PatchesSource.ArchiveFilePath())
	custom.SetSource(config.PatchesSource)

	rep.storage = geonames.NewMultiStorage[geonames.Storage](original).Add(custom)

	go rep.initAutoUpdates(context.Background(), config.AutoUpdatePeriod)
	return rep
}

func (r *GeoNameRepository) CheckUpdates(ctx context.Context) (entity.Updates, error) {
	result, err, _ := r.checkUpdatesSF.Do("check_updates", func() (interface{}, error) {
		updates, err := r.storage.CheckUpdates(ctx)
		return updates, err
	})

	if err != nil {
		return nil, err
	}

	return result.(entity.Updates), nil
}

func (r *GeoNameRepository) Download(ctx context.Context) (entity.Updates, error) {
	if !r.updateM.TryLock() {
		return nil, utils.ErrUpdateInProgress
	}
	defer r.updateM.Unlock()
	return r.storage.Download(ctx)
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

func (r *GeoNameRepository) State() *state.GeosState {
	return r.storage.State()
}

func (r *GeoNameRepository) initAutoUpdates(ctx context.Context, hoursPeriod int) {
	go func() {
		ticker := time.NewTicker(time.Duration(hoursPeriod) * time.Hour)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				upd, err := r.Download(ctx)
				if err != nil {
					log.FromContext(ctx).ErrorfWithFields(log.Fields{"err": err}, "failed to auto-update due to download error")
					continue
				}

				if len(upd) == 0 {
					log.FromContext(ctx).Info("No geoname updates found")
					continue
				}

				for subj, status := range upd {
					if status.Error != "" {
						log.FromContext(ctx).ErrorfWithFields(log.Fields{"err": status.Error}, "failed to auto-update due to download error for %s", subj)
					} else {
						log.FromContext(ctx).Infof("Successfully auto-updated %s", subj)
					}
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}
