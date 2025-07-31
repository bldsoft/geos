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

func Open(source *source.MMDBSource) (*MaxmindDatabase, error) {
	ctx := context.Background()
	res := &MaxmindDatabase{
		source: source,
	}

	version, err := source.Version(ctx)
	if err != nil {
		return nil, err
	}

	if err := res.update(ctx, version); err != nil {
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
		if err := db.update(ctx, update.RemoteVersion); err != nil {
			return err
		}
	}

	return nil
}

func (db *MaxmindDatabase) update(ctx context.Context, sourceVersion source.MMDBVersion) error {
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

	db.lastUpdate = entity.MMDBVersion(sourceVersion)
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
