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

type GeoIPRepository struct {
	dbCity *maxmindDBWithCachedCSVDump
	dbISP  *maxmindDBWithCachedCSVDump
}

func NewGeoIPRepository(dbCityPath, dbISPPath string, csvDirPath string) *GeoIPRepository {
	rep := GeoIPRepository{}

	cityOrigDB, err := maxmind.Open(dbCityPath)
	if err != nil {
		log.Fatalf("Failed to read %s db: %s", dbCityPath, err)
	}
	cityCustomDBs := maxmind.NewCustomDatabasesFromDir(filepath.Dir(dbCityPath), "city")
	patchedCityDB := maxmind.NewMultiMaxMindDB(cityOrigDB).Add(cityCustomDBs...).WithLogger(*log.Logger.WithFields(log.Fields{"db": "city"}))
	cityDB := newCityDB(patchedCityDB)
	rep.dbCity = withCachedCSVDump(cityDB)

	var ispDB maxmind.CSVDumper
	ispOrigDB, err := maxmind.Open(dbISPPath)
	if err != nil {
		log.Warnf("Failed to read %s db: %s", dbISPPath, err)
	} else {
		ispCustomDBs := maxmind.NewCustomDatabasesFromDir(filepath.Dir(dbCityPath), "isp")
		patchedISPDB := maxmind.NewMultiMaxMindDB(ispOrigDB).Add(ispCustomDBs...).WithLogger(*log.Logger.WithFields(log.Fields{"db": "isp"}))
		ispDB = newISPDB(patchedISPDB)
		rep.dbISP = withCachedCSVDump(ispDB)
	}

	go rep.initCSVDumps(csvDirPath)
	return &rep
}

func (r *GeoIPRepository) initCSVDumps(csvDirPath string) {
	ctx := context.Background()
	r.dbCity.initCSVDump(ctx, filepath.Join(csvDirPath, "dump.csv"))
	if r.dbISP != nil {
		r.dbISP.initCSVDump(ctx, filepath.Join(csvDirPath, "isp.csv"))
	} else {
		log.FromContext(ctx).InfoWithFields(log.Fields{"path": csvDirPath}, "Skipping csv dump load")
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
		db = r.dbISP
	default:
		return nil, errors.New("unknown database type")
	}

	if db == nil {
		return nil, ErrDBNotAvailable
	}
	return db, nil
}
