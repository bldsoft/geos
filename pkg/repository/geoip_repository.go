package repository

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"path/filepath"
	"slices"
	"sync"
	"time"

	"github.com/bldsoft/geos/pkg/entity"
	"github.com/bldsoft/geos/pkg/storage/maxmind"
	"github.com/bldsoft/geos/pkg/storage/source"
	"github.com/bldsoft/geos/pkg/utils"
	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/utils/errgroup"
	"go.uber.org/atomic"
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
) *maxmindDBWithCachedCSVDump {
	dbSource := source.NewMMDBSource(conf.RemoteURL, conf.LocalPath)
	originalDB, err := maxmind.Open(dbSource)
	if err != nil {
		if required {
			log.Fatalf("Failed to read %s db: %s", customPrefix, err)
		}
		log.Warnf("Failed to read %s db: %s", customPrefix, err)
		return nil
	}

	patchesSource := source.NewPatchesSource(conf.PatchesRemoteURL, filepath.Dir(conf.LocalPath), customPrefix)
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
	dbCity, dbISP                       *maxmindDBWithCachedCSVDump
	cityConf, ispConf                   DBConfig
	csvDirPath                          string
	checkUpdatesSF                      singleflight.Group
	cityUpdateLastErr, ispUpdateLastErr atomic.Pointer[string]
	fileRep                             source.LocalFileRepository
}

func NewGeoIPRepository(cityConf, ispConf DBConfig, csvDirPath string) *GeoIPRepository {
	return &GeoIPRepository{
		csvDirPath: csvDirPath,
		cityConf:   cityConf,
		ispConf:    ispConf,
		dbCity:     openPatchedDB[entity.City](cityConf, string(MaxmindDBTypeCity), csvDirPath, true),
		dbISP:      openPatchedDB[entity.ISP](ispConf, string(MaxmindDBTypeISP), csvDirPath, false),
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

func (r *GeoIPRepository) CheckUpdates(ctx context.Context) ([]entity.DBUpdate, error) {
	result, err, _ := r.checkUpdatesSF.Do("check_updates", func() (interface{}, error) {
		var multiErr error

		cityUpdate, err := r.dbCity.CheckUpdates(ctx)
		if err != nil {
			multiErr = errors.Join(multiErr, err)
		}

		res := make([]entity.DBUpdate, 0, 2)
		res = append(res, entity.NewDBUpdate(string(MaxmindDBTypeCity), cityUpdate, r.cityUpdateLastErr.Load()))
		if r.dbISP == nil {
			return res, multiErr
		}

		ispUpdate, err := r.dbISP.CheckUpdates(ctx)
		if err != nil {
			multiErr = errors.Join(multiErr, err)
		}

		if multiErr != nil {
			return nil, multiErr
		}

		res = append(res, entity.NewDBUpdate(string(MaxmindDBTypeISP), ispUpdate, r.ispUpdateLastErr.Load()))
		return res, nil
	})

	if err != nil {
		return nil, err
	}
	return result.([]entity.DBUpdate), nil
}

func (r *GeoIPRepository) StartUpdate(ctx context.Context) error {
	ok, close, err := r.fileRep.TryLock(ctx, filepath.Join(r.cityConf.LocalPath, "geoip.lock"))
	if !ok || err != nil {
		if errors.Is(err, source.ErrFileExists) {
			return utils.ErrUpdateInProgress
		}
		return err
	}

	g := errgroup.Group{}
	g.Go(func() error {
		defer close()
		return r.TryUpdate(ctx)
	})
	return g.Wait()
}

func (r *GeoIPRepository) TryUpdate(ctx context.Context) error {
	var multiErr error

	for _, db := range []struct {
		db      *maxmindDBWithCachedCSVDump
		lastErr *atomic.Pointer[string]
	}{
		{db: r.dbCity, lastErr: &r.cityUpdateLastErr},
		{db: r.dbISP, lastErr: &r.ispUpdateLastErr},
	} {
		if db.db == nil {
			continue
		}
		err := db.db.TryUpdate(ctx)
		multiErr = errors.Join(multiErr, err)
		if err == nil {
			db.lastErr.Store(nil)
		} else {
			errStr := err.Error()
			db.lastErr.Store(&errStr)
		}
	}
	return multiErr
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

			err = db.TryUpdate(ctx)
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
			upd, err := r.CheckUpdates(ctx)
			if err != nil {
				log.FromContext(ctx).ErrorfWithFields(log.Fields{
					"err": err,
				}, "failed to check updates")
				continue
			}

			if !slices.ContainsFunc(upd, func(u entity.DBUpdate) bool {
				return u.Update.AvailableVersion != ""
			}) {
				continue
			}

			if err := r.TryUpdate(ctx); err != nil {
				log.FromContext(ctx).ErrorfWithFields(log.Fields{
					"err": err,
				}, "failed to update")
				continue
			}

			log.FromContext(ctx).InfoWithFields(log.Fields{"check": upd}, "Successfully auto-updated")
		case <-ctx.Done():
			return nil
		}
	}
}
