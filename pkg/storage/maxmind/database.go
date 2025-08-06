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

	lastUpdate entity.MMDBVersion
}

func Open(ctx context.Context, source *source.MMDBSource) (*MaxmindDatabase, error) {
	res := &MaxmindDatabase{
		source: source,
	}

	if err := res.update(ctx); err != nil {
		return nil, err
	}
	return res, nil
}

func (db *MaxmindDatabase) Lookup(ctx context.Context, ip net.IP, result interface{}) error {
	return db.reader.Load().Lookup(ip, result)
}

func (db *MaxmindDatabase) RawData(ctx context.Context) (io.Reader, error) {
	return bytes.NewBuffer(*db.dbRaw.Load()), nil
}

func (db *MaxmindDatabase) MetaData(ctx context.Context) (*maxminddb.Metadata, error) {
	return &db.reader.Load().Metadata, nil
}

func (db *MaxmindDatabase) Networks(ctx context.Context, options ...maxminddb.NetworksOption) (*maxminddb.Networks, error) {
	return db.reader.Load().Networks(), nil
}

func (db *MaxmindDatabase) Update(ctx context.Context, force bool) error {
	update, err := db.source.CheckUpdates(ctx)
	if err != nil {
		return err
	}

	if update.RemoteVersion.Compare(update.CurrentVersion) > 0 {
		if err := db.source.Update(ctx, force); err != nil {
			return err
		}
	}

	if update.RemoteVersion.Compare(source.MMDBVersion(db.lastUpdate)) > 0 {
		if err := db.update(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (db *MaxmindDatabase) update(ctx context.Context) error {
	reader, err := db.source.Reader(ctx)
	if err != nil {
		return err
	}

	version, err := db.source.Version(ctx)
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

	db.lastUpdate = entity.MMDBVersion(version)
	return nil
}

func (db *MaxmindDatabase) CheckUpdates(ctx context.Context) (entity.Update[entity.MMDBVersion], error) {
	update, err := db.source.CheckUpdates(ctx)
	if err != nil {
		return entity.Update[entity.MMDBVersion]{}, err
	}
	return entity.Update[entity.MMDBVersion]{
		CurrentVersion: entity.MMDBVersion(db.lastUpdate),
		RemoteVersion:  entity.MMDBVersion(update.RemoteVersion),
	}, nil
}
