package repository

import (
	"context"
	"errors"
	"fmt"
	"io"
	"maps"
	"net"
	"path/filepath"
	"sync"
	"time"

	"github.com/bldsoft/geos/pkg/entity"
	"github.com/bldsoft/geos/pkg/storage/maxmind"
	"github.com/bldsoft/geos/pkg/storage/source"
	"github.com/bldsoft/geos/pkg/storage/state"
	"github.com/bldsoft/geos/pkg/utils"
	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/utils/errgroup"
	"golang.org/x/sync/singleflight"
)

var (
	ErrGeoIPCSVNotReady = fmt.Errorf("geoip csv dump is %w", utils.ErrNotReady)
	ErrGeoIPCSVDisabled = fmt.Errorf("geoip csv dump is %w", utils.ErrDisabled)

	ErrCSVNotSupported = fmt.Errorf("%w: csv format", errors.ErrUnsupported)
	ErrDBNotAvailable  = fmt.Errorf("db %w", utils.ErrNotAvailable)

	csvDumpOnce = sync.Once{}
)

type DumpFormat string

const (
	DumpFormatCSV        DumpFormat = "csv"
	DumpFormatGzippedCSV DumpFormat = "csv.gz"
	DumpFormatMMDB       DumpFormat = "mmdb"
)

type MaxmindDBType string

const (
	MaxmindDBTypeCity MaxmindDBType = "city"
	MaxmindDBTypeISP  MaxmindDBType = "isp"
)

func openPatchedDB[T maxmind.CSVEntity](
	conf DBConfig, customPrefix, csvDumpDir string, required bool,
	subject entity.Subject, // TODO: REMOVE
	patchesSubject entity.Subject, // TODO: REMOVE
) *maxmindDBWithCachedCSVDump {
	dbSource := source.NewMMDBSource(conf.RemoteURL, conf.LocalPath, subject)
	originalDB, err := maxmind.Open(dbSource)
	if err != nil {
		if required {
			log.Fatalf("Failed to read %s db: %s", customPrefix, err)
		}
		log.Warnf("Failed to read %s db: %s", customPrefix, err)
		return nil
	}

	patchesSource := source.NewPatchesSource(conf.PatchesRemoteURL, filepath.Dir(conf.LocalPath), string(patchesSubject), patchesSubject)
	customDB := maxmind.NewCustomDatabaseFromTarGz(patchesSource)

	multiDB := maxmind.NewMultiMaxMindDB[maxmind.Database](originalDB).Add(customDB).WithLogger(
		*log.Logger.WithFields(log.Fields{"db": customPrefix}),
	)
	return withCachedCSVDump[T](multiDB, csvDumpDir)
}

type DBConfig struct {
	LocalPath        string
	RemoteURL        string
	PatchesRemoteURL string
	AutoUpdatePeriod time.Duration
}

type GeoIPRepository struct {
	dbCity, dbISP     *maxmindDBWithCachedCSVDump
	cityConf, ispConf DBConfig
	csvDirPath        string
	checkUpdatesSF    singleflight.Group
}

func NewGeoIPRepository(cityConf, ispConf DBConfig, csvDirPath string) *GeoIPRepository {
	return &GeoIPRepository{
		csvDirPath: csvDirPath,
		cityConf:   cityConf,
		ispConf:    ispConf,
		dbCity:     openPatchedDB[entity.City](cityConf, string(MaxmindDBTypeCity), csvDirPath, true, entity.SubjectCitiesDb, entity.SubjectCitiesDbPatches),
		dbISP:      openPatchedDB[entity.ISP](ispConf, string(MaxmindDBTypeISP), csvDirPath, false, entity.SubjectISPDb, entity.SubjectISPDbPatches),
	}
}

func lookup[T any](db maxmind.Database, ip net.IP) (*T, error) {
	var obj T
	return &obj, db.Lookup(ip, &obj)
}

func (r *GeoIPRepository) Country(ctx context.Context, ip net.IP) (*entity.Country, error) {
	return lookup[entity.Country](r.dbCity, ip)
}

func (r *GeoIPRepository) City(ctx context.Context, ip net.IP, includeISP bool) (*entity.City, error) {
	city, err := lookup[entity.City](r.dbCity, ip)
	if err != nil {
		return nil, err
	}
	if includeISP {
		var isp entity.ISP
		err := r.dbISP.Lookup(ip, &isp)
		if err != nil {
			log.FromContext(ctx).ErrorWithFields(log.Fields{"err": err}, "Failed to fill ISP")
		} else {
			city.ISP = &isp
		}
	}
	return city, nil
}

func (r *GeoIPRepository) CityLite(ctx context.Context, ip net.IP, lang string) (*entity.CityLite, error) {
	cityLiteDB, err := lookup[entity.CityLiteDb](r.dbCity, ip)
	if err != nil {
		return nil, err
	}
	return entity.DbToCityLite(cityLiteDB, lang), nil
}

func (r *GeoIPRepository) MetaData(ctx context.Context, dbType MaxmindDBType) (*entity.MetaData, error) {
	db, err := r.database(ctx, dbType)
	if err != nil {
		return nil, err
	}
	return db.MetaData()
}

func (r *GeoIPRepository) Database(ctx context.Context, dbType MaxmindDBType, format DumpFormat) (*entity.Database, error) {
	db, err := r.database(ctx, dbType)
	if err != nil {
		return nil, err
	}
	meta, err := db.MetaData()
	if err != nil {
		return nil, err
	}
	var data io.Reader
	ext := string(format)
	switch format {
	case DumpFormatCSV:
		fallthrough
	case DumpFormatGzippedCSV:
		if csvDumper, ok := db.(maxmind.CSVDumper); ok {
			data, err = csvDumper.CSV(ctx, format == DumpFormatGzippedCSV)
			ext = string(DumpFormatCSV)
		} else {
			return nil, fmt.Errorf("%s: %w", dbType, ErrCSVNotSupported)
		}
	case DumpFormatMMDB:
		data, err = db.RawData()
	default:
		return nil, utils.ErrUnknownFormat
	}
	if err != nil {
		return nil, err
	}
	return &entity.Database{
		Data:     data,
		MetaData: *meta,
		Ext:      "." + ext,
	}, nil
}

func (r *GeoIPRepository) database(ctx context.Context, dbType MaxmindDBType) (maxmind.Database, error) {
	var db maxmind.Database
	switch dbType {
	case MaxmindDBTypeCity:
		db = r.dbCity
	case MaxmindDBTypeISP:
		if r.dbISP == nil {
			return nil, ErrGeoIPCSVDisabled
		}
		db = r.dbISP
	default:
		return nil, errors.New("unknown database type")
	}

	if db == nil {
		return nil, ErrDBNotAvailable
	}
	return db, nil
}

func (r *GeoIPRepository) CheckUpdates(ctx context.Context) (entity.Updates, error) {
	result, err, _ := r.checkUpdatesSF.Do("check_updates", func() (interface{}, error) {
		var multiErr error
		multiUpdates := entity.Updates{}

		updates, err := r.dbCity.CheckUpdates(ctx)
		if err != nil {
			multiErr = errors.Join(multiErr, err)
		}
		if updates != nil {
			maps.Copy(multiUpdates, updates)
		}

		if r.dbISP == nil {
			return multiUpdates, multiErr
		}

		updates, err = r.dbISP.CheckUpdates(ctx)
		if err != nil {
			multiErr = errors.Join(multiErr, err)
		}
		if updates != nil {
			maps.Copy(multiUpdates, updates)
		}

		return multiUpdates, multiErr
	})

	if err != nil {
		return nil, err
	}

	return result.(entity.Updates), nil
}

func (r *GeoIPRepository) Download(ctx context.Context) (entity.Updates, error) {
	multiUpdates := entity.Updates{}
	var resultsMtx sync.Mutex

	var eg errgroup.Group
	for _, db := range []*maxmindDBWithCachedCSVDump{r.dbCity, r.dbISP} {
		if db == nil {
			continue
		}
		eg.Go(func() error {
			updates, err := db.Download(ctx)
			if err != nil {
				return err
			}

			resultsMtx.Lock()
			defer resultsMtx.Unlock()
			maps.Copy(multiUpdates, updates)
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return multiUpdates, nil
}

func (r *GeoIPRepository) State() *state.GeosState {
	result := &state.GeosState{}

	if cityState := r.dbCity.State(); cityState != nil {
		result.Add(cityState)
	}

	if r.dbISP != nil {
		if ispState := r.dbISP.State(); ispState != nil {
			result.Add(ispState)
		}
	}

	return result
}

func (r *GeoIPRepository) Run(ctx context.Context) error {
	for _, db := range []*maxmindDBWithCachedCSVDump{r.dbCity, r.dbISP} {
		if db == nil {
			continue
		}
		logger := log.FromContext(ctx).WithFields(log.Fields{"db": db.metadata().DatabaseType})
		ctx = context.WithValue(ctx, log.LoggerCtxKey, logger)

		ticker := time.NewTicker(time.Minute)
		for {
			interrupted, err := db.lastUpdateInterrupted(ctx)
			if err != nil {
				logger.ErrorfWithFields(log.Fields{
					"err": err,
				}, "failed to check if update was interrupted")
				continue
			}
			if !interrupted {
				continue
			}

			logger.Infof("Found interrupted update, retrying...")

			_, err = db.Download(ctx)
			if err == nil {
				logger.Infof("Successfully updated %s", db.metadata().DatabaseType)
				break
			}
			logger.ErrorfWithFields(log.Fields{
				"err": err,
			}, "failed to update")

			select {
			case <-ticker.C:
			case <-ctx.Done():
				return nil
			}
		}
	}

	if r.cityConf.AutoUpdatePeriod <= 0 {
		return nil
	}

	for {
		select {
		case <-time.After(r.cityConf.AutoUpdatePeriod):
			upd, err := r.Download(ctx)
			if err != nil {
				log.FromContext(ctx).ErrorfWithFields(log.Fields{
					"err": err,
				}, "failed to update")
			}
			if len(upd) == 0 {
				log.FromContext(ctx).Info("No geoip updates found")
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
			return nil
		}
	}
}
