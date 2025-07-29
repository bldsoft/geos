package maxmind

import (
	"bytes"
	"context"
	"io"
	"net"
	"sync/atomic"

	"github.com/bldsoft/geos/pkg/entity"
	"github.com/bldsoft/geos/pkg/storage/source"
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

func (db *MaxmindDatabase) Update(ctx context.Context, force bool) error {
	update, err := db.CheckUpdates(ctx)
	if err != nil {
		return err
	}

	if !update.RemoteVersion.IsHigher(update.CurrentVersion) {
		return nil
	}

	if err := db.source.Update(ctx, force); err != nil {
		return err
	}

	return db.update(ctx)
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

func (db *MaxmindDatabase) CheckUpdates(ctx context.Context) (entity.Update[entity.MMDBVersion], error) {
	return db.source.CheckUpdates(ctx)
}
