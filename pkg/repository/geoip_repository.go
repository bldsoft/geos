package repository

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"path/filepath"
	"sync"
	"time"

	"github.com/bldsoft/geos/pkg/entity"
	"github.com/bldsoft/geos/pkg/storage/maxmind"
	"github.com/bldsoft/geos/pkg/storage/source"
	"github.com/bldsoft/geos/pkg/utils"
	"github.com/bldsoft/gost/log"
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

	multiDB := maxmind.NewMultiMaxMindDB[maxmind.Database](originalDB).WithLogger(
		*log.Logger.WithFields(log.Fields{"db": customPrefix}),
	)

	if conf.PatchesRemoteURL != "" {
		patchesSource := source.NewPatchesSource(conf.PatchesRemoteURL, filepath.Dir(conf.LocalPath), customPrefix)
		customDB := maxmind.NewCustomDatabaseFromTarGz(patchesSource)
		multiDB = multiDB.Add(customDB)
	}

	return withCachedCSVDump[T](multiDB, csvDumpDir)
}

type DBConfig struct {
	LocalPath        string
	RemoteURL        string
	PatchesRemoteURL string
}

type GeoIPRepositoryConfig struct {
	City             DBConfig
	ISP              DBConfig
	CSVDirPath       string
	AutoUpdatePeriod time.Duration
}

type GeoIPRepository struct {
	cfg            GeoIPRepositoryConfig
	dbCity, dbISP  *maxmindDBWithCachedCSVDump
	csvDirPath     string
	checkUpdatesSF singleflight.Group

	cityUpdater, ispUpdater *updaterWithLastErr

	*baseUpdateRepository
}

func NewGeoIPRepository(cfg GeoIPRepositoryConfig) *GeoIPRepository {
	res := &GeoIPRepository{
		cfg:    cfg,
		dbCity: openPatchedDB[entity.City](cfg.City, string(MaxmindDBTypeCity), cfg.CSVDirPath, true),
		dbISP:  openPatchedDB[entity.ISP](cfg.ISP, string(MaxmindDBTypeISP), cfg.CSVDirPath, false),
	}
	res.cityUpdater = newUpdaterWithLastErr(res.dbCity.Update)
	res.ispUpdater = newUpdaterWithLastErr(res.dbISP.Update)
	res.baseUpdateRepository = NewBaseUpdateRepository(
		"geoip.lock",
		cfg.AutoUpdatePeriod,
		func(ctx context.Context, force bool) error {
			err := res.cityUpdater.Update(ctx, force)
			if err != nil {
				return err
			}
			if res.dbISP != nil {
				err := res.ispUpdater.Update(ctx, force)
				if err != nil {
					return err
				}
			}
			return nil
		},
	)
	return res
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
		res = append(res, entity.NewDBUpdate(
			string(MaxmindDBTypeCity),
			cityUpdate,
			r.cityUpdater.InProgress(),
			r.cityUpdater.LastErr(),
		))
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

		res = append(res, entity.NewDBUpdate(
			string(MaxmindDBTypeISP),
			ispUpdate,
			r.ispUpdater.InProgress(),
			r.ispUpdater.LastErr(),
		))
		return res, nil
	})

	if err != nil {
		return nil, err
	}
	return result.([]entity.DBUpdate), nil
}
