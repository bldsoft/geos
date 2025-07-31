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

func NewCustomDatabase(source *source.TSUpdatableFile) *CustomDatabase {
	logger := log.Logger.WithFields(log.Fields{"source": source.LocalPath, "db": "custom maxmind"})

	res := &CustomDatabase{
		source: source,
	}
	res.base.Store(NewMultiMaxMindDB())

	version, err := source.Version(context.Background())
	if err != nil {
		logger.Errorf("failed to get local version: %v", err)
	}

	if err := res.update(context.Background(), version); err != nil {
		logger.Errorf("failed to get patches: %v", err)
	}

	return res
}

func (db *CustomDatabase) db() *MultiMaxMindDB {
	return db.base.Load()
}

func (db *CustomDatabase) Lookup(ip net.IP, result interface{}) error {
	return db.db().Lookup(ip, result)
}

func (db *CustomDatabase) Networks(options ...maxminddb.NetworksOption) (*maxminddb.Networks, error) {
	return db.db().Networks(options...)
}

func (db *CustomDatabase) RawData() (io.Reader, error) {
	return db.db().RawData()
}

func (db *CustomDatabase) MetaData() (*maxminddb.Metadata, error) {
	return db.db().MetaData()
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
