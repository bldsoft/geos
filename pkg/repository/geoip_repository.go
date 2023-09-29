package repository

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"path/filepath"

	"github.com/bldsoft/geos/pkg/entity"
	"github.com/bldsoft/geos/pkg/utils"
	"github.com/bldsoft/gost/log"
	"github.com/oschwald/maxminddb-golang"
)

var (
	ErrGeoIPCSVNotReady = fmt.Errorf("geoip csv dump is %w", utils.ErrNotReady)
	ErrGeoIPCSVDisabled = fmt.Errorf("geoip csv dump is %w", utils.ErrDisabled)
)

type DumpFormat string

const (
	DumpFormatCSV          DumpFormat = "csv"
	DumpFormatCSVWithNames DumpFormat = "csvWithNames"
	DumpFormatMMDB         DumpFormat = "mmdb"
)

type MaxmindDBType string

const (
	MaxmindDBTypeCity MaxmindDBType = "city"
	MaxmindDBTypeISP  MaxmindDBType = "isp"
)

type maxmindDatabase interface {
	Lookup(ip net.IP, result interface{}) error
	LookupNetwork(ip net.IP, result interface{}) (network *net.IPNet, ok bool, err error)
	LookupOffset(ip net.IP) (uintptr, error)
	Networks(options ...maxminddb.NetworksOption) (*maxminddb.Networks, error)
	NetworksWithin(network *net.IPNet, options ...maxminddb.NetworksOption) (*maxminddb.Networks, error)
	Verify() error
	Close() error

	Available() bool
	Path() (string, error)
	RawData() (io.Reader, error) // mmdb
	MetaData() (*maxminddb.Metadata, error)
}
type csvDumper interface {
	WriteCSVTo(ctx context.Context, w io.Writer) error
	CSV(ctx context.Context, withColumnNames bool) (io.Reader, error)
}

type maxmindCSVDumper interface {
	maxmindDatabase
	csvDumper
}

type GeoIpRepository struct {
	dbCity *maxmindDBWithCachedCSVDump
	dbISP  *maxmindDBWithCachedCSVDump
}

func NewGeoIpRepository(dbCityPath, dbISPPath string, csvDirPath string) *GeoIpRepository {
	rep := &GeoIpRepository{
		dbCity: withCachedCSVDump(openCityDB(dbCityPath, true)),
		dbISP:  withCachedCSVDump(openISPDB(dbISPPath, false)),
	}

	go rep.initCSVDumps(csvDirPath)
	return rep
}

func (r *GeoIpRepository) initCSVDumps(csvDirPath string) {
	ctx := context.Background()
	r.dbCity.initCSVDump(ctx, filepath.Join(csvDirPath, "dump.csv"))
	r.dbISP.initCSVDump(ctx, filepath.Join(csvDirPath, "isp.csv"))

}

func lookup[T any](db maxmindDatabase, ip net.IP) (*T, error) {
	var obj T
	return &obj, db.Lookup(ip, &obj)
}

func (r *GeoIpRepository) Country(ctx context.Context, ip net.IP) (*entity.Country, error) {
	return lookup[entity.Country](r.dbCity, ip)
}

func (r *GeoIpRepository) City(ctx context.Context, ip net.IP, includeISP bool) (*entity.City, error) {
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

func (r *GeoIpRepository) CityLite(ctx context.Context, ip net.IP, lang string) (*entity.CityLite, error) {
	cityLiteDb, err := lookup[entity.CityLiteDb](r.dbCity, ip)
	if err != nil {
		return nil, err
	}
	return entity.DbToCityLite(cityLiteDb, lang), nil
}

func (r *GeoIpRepository) Database(ctx context.Context, dbType MaxmindDBType, format DumpFormat) (*entity.Database, error) {
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
	case DumpFormatCSVWithNames:
		if csvDumper, ok := db.(csvDumper); ok {
			data, err = csvDumper.CSV(ctx, format == DumpFormatCSVWithNames)
			ext = string(DumpFormatCSV)
		} else {
			return nil, fmt.Errorf("csv format for %s is not supported", dbType)
		}
	case DumpFormatMMDB:
		data, err = db.RawData()
	default:
		return nil, errors.New("unknown format")
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

func (r *GeoIpRepository) database(ctx context.Context, dbType MaxmindDBType) (maxmindDatabase, error) {
	switch dbType {
	case MaxmindDBTypeCity:
		return r.dbCity, nil
	case MaxmindDBTypeISP:
		return r.dbISP, nil
	default:
		return nil, errors.New("Unknown database type")
	}
}
