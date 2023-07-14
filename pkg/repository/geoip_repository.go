package repository

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strconv"

	"github.com/bldsoft/geos/pkg/entity"
	"github.com/bldsoft/gost/log"
	"github.com/oschwald/maxminddb-golang"
)

var (
	ErrGeoIPCSVNotReady = errors.New("geoip csv dump not ready")
	ErrGeoIPCSVDisabled = errors.New("geoip csv dump is disabled")
)

type DumpFormat string

const (
	DumpFormatCSV          DumpFormat = "csv"
	DumpFormatCSVWithNames DumpFormat = "csvWithNames"
)

type MaxmindDBType string

const (
	City MaxmindDBType = "city"
)

type GeoIpRepository struct {
	dbRaw            []byte
	db               *maxminddb.Reader
	dumpReady        chan struct{}
	csvWithNamesDump []byte
}

func NewGeoIpRepository(dbPath string, csvDirPath string) *GeoIpRepository {
	dbRaw, err := ioutil.ReadFile(dbPath)
	if err != nil {
		log.Fatalf("Failed to read db: %s", err)
	}

	db, err := maxminddb.FromBytes(dbRaw)
	rep := &GeoIpRepository{dbRaw: dbRaw, db: db, dumpReady: make(chan struct{})}

	go func() {
		defer close(rep.dumpReady)
		if csvDirPath == "" {
			return
		}

		var err error
		ctx := context.Background()

		if err := os.MkdirAll(csvDirPath, os.ModePerm); err != nil {
			panic(fmt.Errorf("failed to create dir for csv dump: %w", err))
		}

		dumpPath := filepath.Join(csvDirPath, "dump.csv")
		rep.csvWithNamesDump, err = rep.getDumpFromDisk(ctx, dbPath, dumpPath)
		if err != nil {
			panic(fmt.Errorf("failed to load GeoIP dump: %w", err))
		}
		log.FromContext(ctx).Infof("Dump loaded to memory, size %d MB", len(rep.csvWithNamesDump)/1024/1024)
	}()
	return rep
}

func lookup[T any](db *maxminddb.Reader, ip net.IP) (*T, error) {
	var obj T
	return &obj, db.Lookup(ip, &obj)
}

func (r *GeoIpRepository) Country(ctx context.Context, ip net.IP) (*entity.Country, error) {
	return lookup[entity.Country](r.db, ip)
}

func (r *GeoIpRepository) City(ctx context.Context, ip net.IP) (*entity.City, error) {
	return lookup[entity.City](r.db, ip)
}

func (r *GeoIpRepository) CityLite(ctx context.Context, ip net.IP, lang string) (*entity.CityLite, error) {
	cityLiteDb, err := lookup[entity.CityLiteDb](r.db, ip)
	if err != nil {
		return nil, err
	}
	return entity.DbToCityLite(cityLiteDb, lang), nil
}

func (r *GeoIpRepository) Database(ctx context.Context, dbType MaxmindDBType) (*entity.Database, error) {
	switch dbType {
	case City:
		return &entity.Database{
			Data:     r.dbRaw,
			MetaData: r.db.Metadata,
			Ext:      ".mmdb",
		}, nil
	default:
		return nil, errors.New("Unknown database type")
	}
}

func (r *GeoIpRepository) Dump(ctx context.Context, format DumpFormat) ([]byte, error) {
	select {
	case <-r.dumpReady:
		if r.csvWithNamesDump == nil {
			return nil, ErrGeoIPCSVDisabled
		}
		if format == DumpFormatCSV {
			return r.removeFirstLine(r.csvWithNamesDump), nil
		}
		return r.csvWithNamesDump, nil
	default:
		return nil, ErrGeoIPCSVNotReady
	}
}

func (r *GeoIpRepository) removeFirstLine(buf []byte) []byte {
	i := bytes.Index(r.csvWithNamesDump, []byte("\n"))
	if i == -1 {
		return nil
	}
	return r.csvWithNamesDump[i+1:]
}

func (r *GeoIpRepository) getDumpFromDisk(ctx context.Context, dbPath, dumpPath string) ([]byte, error) {
	needUpdate, err := r.needUpdateDump(ctx, dbPath, dumpPath)
	if err != nil {
		return nil, err
	}
	if !needUpdate {
		return ioutil.ReadFile(dumpPath)
	}

	log.FromContext(ctx).InfoWithFields(log.Fields{"db": dbPath, "csv": dumpPath}, "Updating CSV")
	return r.loadDumpFull(ctx, dbPath, dumpPath)
}

func (r *GeoIpRepository) needUpdateDump(ctx context.Context, dbPath, dumpPath string) (bool, error) {
	dbStat, err := os.Stat(dbPath)
	if err != nil {
		return false, err
	}
	dumpStat, err := os.Stat(dumpPath)
	if err != nil {
		if os.IsNotExist(err) {
			return true, nil
		}
		return false, err
	}
	return dbStat.ModTime().After(dumpStat.ModTime()), nil
}

func (r *GeoIpRepository) loadDumpFull(ctx context.Context, dbPath, dumpPath string) ([]byte, error) {
	temp := dumpPath + ".tmp"
	dump, err := r.loadDump(ctx, dbPath, temp)
	if err != nil {
		return nil, os.Remove(temp)
	}
	return dump, os.Rename(temp, dumpPath)
}

func (r *GeoIpRepository) loadDump(ctx context.Context, dbPath, dumpPath string) ([]byte, error) {
	networks := r.db.Networks(maxminddb.SkipAliasedNetworks)

	var buf bytes.Buffer
	file, err := os.Create(dumpPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	csvWriter := csv.NewWriter(io.MultiWriter(&buf, file))

	if err := csvWriter.Write([]string{
		"network",
		"city_geoname_id",
		"subdivision_geoname_id",
		"registered_country_geoname_id",
		"represented_country_geoname_id",
		"is_anonymous_proxy",
		"is_satelite_provider",
		"latitude",
		"longitude",
		"accuracy_radius",
	}); err != nil {
		return nil, err
	}

	var record entity.City
	for networks.Next() {
		subnet, err := networks.Network(&record)
		if err != nil {
			return nil, err
		}

		var subdivisionGeonameID uint64
		if len(record.Subdivisions) > 0 {
			subdivisionGeonameID = uint64(record.Subdivisions[0].GeoNameID)
		}
		err = csvWriter.Write([]string{
			subnet.String(),
			strconv.FormatUint(uint64(record.City.GeoNameID), 10),
			strconv.FormatUint(subdivisionGeonameID, 10),
			strconv.FormatUint(uint64(record.RegisteredCountry.GeoNameID), 10),
			strconv.FormatUint(uint64(record.RepresentedCountry.GeoNameID), 10),
			formatBool(record.Traits.IsAnonymousProxy),
			formatBool(record.Traits.IsSatelliteProvider),
			strconv.FormatFloat(record.Location.Latitude, 'f', 4, 64),
			strconv.FormatFloat(record.Location.Longitude, 'f', 4, 64),
			strconv.FormatUint(uint64(record.Location.AccuracyRadius), 10),
		})
		if err != nil {
			return nil, err
		}
	}

	csvWriter.Flush()
	return buf.Bytes(), nil
}

func formatBool(b bool) string {
	if b {
		return "1"
	}
	return "0"
}
