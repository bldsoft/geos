package repository

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"path/filepath"

	"github.com/bldsoft/geos/pkg/entity"
	"github.com/bldsoft/geos/pkg/storage/maxmind"
	"github.com/bldsoft/geos/pkg/utils"
	"github.com/bldsoft/gost/log"
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

func openPatchedDB[T maxmind.CSVEntity](path string, customPrefix string, required bool) *maxmindDBWithCachedCSVDump {
	originalDB, err := maxmind.Open(path)
	if err != nil {
		if required {
			log.Fatalf("Failed to read %s db: %s", customPrefix, err)
		}
		log.Warnf("Failed to read %s db: %s", customPrefix, err)
		return nil
	}
	customDBs := maxmind.NewCustomDatabasesFromDir(filepath.Dir(path), customPrefix)
	patchedDB := maxmind.NewMultiMaxMindDB(originalDB).Add(customDBs...)
	patchedDB = patchedDB.WithLogger(*log.Logger.WithFields(log.Fields{"db": customPrefix}))
	return withCachedCSVDump[T](patchedDB)
}

type GeoIPRepository struct {
	dbCity  *maxmindDBWithCachedCSVDump
	dbISP   *maxmindDBWithCachedCSVDump
	ispPath string
}

func NewGeoIPRepository(dbCityPath, dbISPPath string, csvDirPath string) *GeoIPRepository {
	rep := GeoIPRepository{
		dbCity:  openPatchedDB[entity.City](dbCityPath, string(MaxmindDBTypeCity), true),
		dbISP:   openPatchedDB[entity.ISP](dbISPPath, string(MaxmindDBTypeISP), false),
		ispPath: dbISPPath,
	}

	go rep.initCSVDumps(context.Background(), csvDirPath)
	return &rep
}

func (r *GeoIPRepository) initCSVDumps(ctx context.Context, csvDirPath string) {
	r.dbCity.initCSVDump(ctx, filepath.Join(csvDirPath, "dump.csv"), filepath.Join(csvDirPath, "cities_gen.meta"))

	if r.dbISP != nil {
		r.dbISP.initCSVDump(ctx, filepath.Join(csvDirPath, "isp.csv"), filepath.Join(csvDirPath, "isp_gen.meta"))
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
