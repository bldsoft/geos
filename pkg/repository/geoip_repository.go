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

	"github.com/bldsoft/geos/pkg/entity"
	"github.com/bldsoft/geos/pkg/storage/maxmind"
	"github.com/bldsoft/geos/pkg/storage/source"
	"github.com/bldsoft/geos/pkg/utils"
	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/utils/errgroup"
)

var (
	ErrGeoIPCSVNotReady = fmt.Errorf("geoip csv dump is %w", utils.ErrNotReady)
	ErrGeoIPCSVDisabled = fmt.Errorf("geoip csv dump is %w", utils.ErrDisabled)

	ErrCSVNotSupported = fmt.Errorf("%w: csv format", errors.ErrUnsupported)
	ErrDBNotAvailable  = fmt.Errorf("db %w", utils.ErrNotAvailable)
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

func openPatchedDB[T maxmind.CSVEntity](conf *DBConfig, customPrefix string, required bool) *maxmindDBWithCachedCSVDump {
	originalDB, err := maxmind.Open(conf.Path)
	if err != nil {
		if required {
			log.Fatalf("Failed to read %s db: %s", customPrefix, err)
		}
		log.Warnf("Failed to read %s db: %s", customPrefix, err)
		return nil
	}
	originalDB.SetSource(conf.DBSource)

	customDB := maxmind.NewCustomDatabaseFromDir(filepath.Dir(conf.Path), customPrefix)
	customDB.SetSource(conf.PatchesSource)

	multiDB := maxmind.NewMultiMaxMindDB[maxmind.Database](originalDB).Add(customDB).WithLogger(
		*log.Logger.WithFields(log.Fields{"db": customPrefix}),
	)
	return withCachedCSVDump[T](multiDB)
}

type DBConfig struct {
	Path          string
	DBSource      *source.MaxmindSource
	PatchesSource *source.PatchesSource
}

type GeoIPRepository struct {
	dbCity, dbISP       *maxmindDBWithCachedCSVDump
	cityConf, ispConf   *DBConfig
	csvDirPath, ispPath string
}

func NewGeoIPRepository(cityConf, ispConf *DBConfig, csvDirPath string) *GeoIPRepository {
	rep := GeoIPRepository{
		csvDirPath: csvDirPath,
		cityConf:   cityConf,
		ispConf:    ispConf,
		dbCity:     openPatchedDB[entity.City](cityConf, string(MaxmindDBTypeCity), true),
		dbISP:      openPatchedDB[entity.ISP](ispConf, string(MaxmindDBTypeISP), false),
		ispPath:    ispConf.Path,
	}

	go rep.initCSVDumps(context.Background(), csvDirPath)
	return &rep
}

func (r *GeoIPRepository) initCSVDumps(ctx context.Context, csvDirPath string) {
	r.dbCity.initCSVDump(ctx, filepath.Join(csvDirPath, "dump.csv"))

	if r.dbISP != nil {
		r.dbISP.initCSVDump(ctx, filepath.Join(csvDirPath, "isp.csv"))
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
		if len(r.ispPath) == 0 {
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
}

func (r *GeoIPRepository) Download(ctx context.Context, update ...bool) (entity.Updates, error) {
	multiUpdates := entity.Updates{}
	var updatesM sync.Mutex

	var eg errgroup.Group
	eg.Go(func() error {
		updates, err := r.dbCity.Download(ctx, update...)
		if err != nil {
			return err
		}

		updatesM.Lock()
		defer updatesM.Unlock()
		maps.Copy(multiUpdates, updates)
		return nil
	})

	eg.Go(func() error {
		if r.dbISP == nil {
			return nil
		}

		updates, err := r.dbISP.Download(ctx, update...)
		if err != nil {
			return err
		}

		updatesM.Lock()
		defer updatesM.Unlock()
		maps.Copy(multiUpdates, updates)
		return nil
	})

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	go func() {
		r.dbCity = openPatchedDB[entity.City](r.cityConf, string(MaxmindDBTypeCity), true)
		r.dbISP = openPatchedDB[entity.ISP](r.ispConf, string(MaxmindDBTypeISP), false)

		go r.initCSVDumps(ctx, r.csvDirPath)
	}()

	return multiUpdates, nil
}
