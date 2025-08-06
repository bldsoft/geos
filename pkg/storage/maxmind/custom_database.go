package maxmind

import (
	"context"
	"io"
	"net"
	"path/filepath"
	"sync/atomic"

	"github.com/bldsoft/geos/pkg/storage/source"
	"github.com/bldsoft/gost/log"
	"github.com/oschwald/maxminddb-golang"
)

type CustomDatabase struct {
	base       atomic.Pointer[MultiMaxMindDB]
	source     *source.TSUpdatableFile
	lastUpdate source.ModTimeVersion
}

func NewCustomDatabase(ctx context.Context, source *source.TSUpdatableFile) *CustomDatabase {
	ctx = context.WithValue(ctx, log.LoggerCtxKey, log.FromContext(ctx).WithFields(log.Fields{"type": "patch"}))

	res := &CustomDatabase{
		source: source,
	}
	res.base.Store(NewMultiMaxMindDB())

	version, err := source.Version(ctx)
	if err != nil {
		log.FromContext(ctx).Errorf("Failed to get local version: %v", err)
		return res
	}

	if err := res.update(ctx, version); err != nil {
		log.FromContext(ctx).Errorf("Failed to get patches: %v", err)
	}

	return res
}

func (db *CustomDatabase) db() *MultiMaxMindDB {
	return db.base.Load()
}

func (db *CustomDatabase) Lookup(ctx context.Context, ip net.IP, result interface{}) error {
	return db.db().Lookup(ctx, ip, result)
}

func (db *CustomDatabase) Networks(ctx context.Context, options ...maxminddb.NetworksOption) (*maxminddb.Networks, error) {
	return db.db().Networks(ctx, options...)
}

func (db *CustomDatabase) RawData(ctx context.Context) (io.Reader, error) {
	return db.db().RawData(ctx)
}

func (db *CustomDatabase) MetaData(ctx context.Context) (*maxminddb.Metadata, error) {
	return db.db().MetaData(ctx)
}

func (db *CustomDatabase) Update(ctx context.Context, force bool) error {
	update, err := db.source.CheckUpdates(ctx)
	if err != nil {
		return err
	}

	if update.RemoteVersion.Compare(update.CurrentVersion) > 0 {
		if err := db.source.Update(ctx, force); err != nil {
			return err
		}
	}

	if update.RemoteVersion.Compare(db.lastUpdate) > 0 {
		if err := db.update(ctx, update.RemoteVersion); err != nil {
			return err
		}
	}
	return nil
}

func (db *CustomDatabase) update(_ context.Context, version source.ModTimeVersion) error {
	var patches []Database
	var err error
	if filepath.Ext(db.source.LocalPath) == ".json" {
		patch, err := NewDatabasePatchFromJSON(db.source)
		if err != nil {
			return err
		}
		patches = []Database{patch}
	} else {
		patches, err = NewDatabasePatchesFromTarGz(db.source)
		if err != nil {
			return err
		}
	}
	db.base.Store(NewMultiMaxMindDB(patches...))
	db.lastUpdate = version
	return nil
}

func (db *CustomDatabase) CheckUpdates(ctx context.Context) (source.Update[source.ModTimeVersion], error) {
	update, err := db.source.CheckUpdates(ctx)
	if err != nil {
		return source.Update[source.ModTimeVersion]{}, err
	}
	update.CurrentVersion = db.lastUpdate
	return update, nil
}
