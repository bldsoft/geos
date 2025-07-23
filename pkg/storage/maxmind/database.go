package maxmind

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"sync/atomic"

	"github.com/bldsoft/geos/pkg/entity"
	"github.com/bldsoft/geos/pkg/storage/source"
	"github.com/bldsoft/geos/pkg/storage/state"
	"github.com/oschwald/maxminddb-golang"
)

type MaxmindDatabase struct {
	source *source.MMDBSource

	reader atomic.Pointer[maxminddb.Reader]
	dbRaw  atomic.Pointer[[]byte]
}

func Open(source *source.MMDBSource) (*MaxmindDatabase, error) {
	ctx := context.Background()
	res := &MaxmindDatabase{
		source: source,
	}
	if err := res.update(ctx); err != nil {
		return nil, err
	}
	return res, nil
}

func (db *MaxmindDatabase) State() *state.GeosState {
	meta := db.reader.Load().Metadata
	versionString := fmt.Sprintf("%d.%d", meta.BinaryFormatMajorVersion, meta.BinaryFormatMinorVersion)
	result := new(state.GeosState)

	switch meta.DatabaseType {
	case "GeoIP2-City", "GeoLite2-City":
		result.CityVersion = versionString
	case "GeoIP2-ISP", "GeoLite2-ISP":
		result.ISPVersion = versionString
	default:
		result.CityVersion = versionString
	}

	return result
}

func (db *MaxmindDatabase) Lookup(ip net.IP, result interface{}) error {
	return db.reader.Load().Lookup(ip, result)
}

func (db *MaxmindDatabase) RawData() (io.Reader, error) {
	return bytes.NewBuffer(*db.dbRaw.Load()), nil
}

func (db *MaxmindDatabase) MetaData() (*maxminddb.Metadata, error) {
	return &db.reader.Load().Metadata, nil
}

func (db *MaxmindDatabase) Networks(options ...maxminddb.NetworksOption) (*maxminddb.Networks, error) {
	return db.reader.Load().Networks(), nil
}

func (db *MaxmindDatabase) Download(ctx context.Context) (entity.Updates, error) {
	updates, err := db.source.Download(ctx)
	if err != nil {
		return nil, err
	}

	if err := db.update(ctx); err != nil {
		return nil, err
	}

	return updates, nil
}

func (db *MaxmindDatabase) update(ctx context.Context) error {
	reader, err := db.source.Reader(ctx)
	if err != nil {
		return err
	}

	dbRaw, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	dbReader, err := maxminddb.FromBytes(dbRaw)
	if err != nil {
		return err
	}

	db.reader.Store(dbReader)
	db.dbRaw.Store(&dbRaw)

	return nil
}

func (db *MaxmindDatabase) CheckUpdates(ctx context.Context) (entity.Updates, error) {
	return db.source.CheckUpdates(ctx)
}

func (db *MaxmindDatabase) LastUpdateInterrupted(ctx context.Context) (bool, error) {
	return db.source.LastUpdateInterrupted(ctx)
}
