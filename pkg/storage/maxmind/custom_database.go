package maxmind

import (
	"context"

	"github.com/bldsoft/geos/pkg/entity"
	"github.com/bldsoft/geos/pkg/storage/source"
	"github.com/bldsoft/gost/log"
	"github.com/oschwald/maxminddb-golang"
)

type CustomDatabase struct {
	*MultiMaxMindDB[*DatabasePatch]
	source       *source.PatchesSource
	customDBName string
}

func NewCustomDatabase(customDBName string, patches ...*DatabasePatch) *CustomDatabase {
	return &CustomDatabase{
		customDBName:   customDBName,
		MultiMaxMindDB: &MultiMaxMindDB[*DatabasePatch]{dbs: patches, logger: log.Logger},
	}
}

func NewCustomDatabaseFromDir(dir, customDBPrefix string) *CustomDatabase {
	customDBs := NewDatabasePatchesFromDir(dir, customDBPrefix)
	return NewCustomDatabase(customDBPrefix, customDBs...)
}

func (db *CustomDatabase) SetSource(source *source.PatchesSource) {
	db.source = source
}

func (db *CustomDatabase) MetaData() (*maxminddb.Metadata, error) {
	if len(db.dbs) == 0 {
		return &maxminddb.Metadata{}, nil //there is no problem if there are no patches
	}

	return db.MultiMaxMindDB.MetaData()
}

func (db *CustomDatabase) Download(ctx context.Context, update ...bool) (updates entity.Updates, err error) {
	if db.source == nil {
		return nil, ErrNoSource
	}

	if updates, err = db.source.Download(ctx, update...); err != nil {
		return nil, err
	}

	db.dbs = NewCustomDatabaseFromDir(db.source.DirPath(), db.customDBName).dbs
	return
}

func (db *CustomDatabase) CheckUpdates(ctx context.Context) (entity.Updates, error) {
	if db.source == nil {
		return nil, ErrNoSource
	}
	return db.source.CheckUpdates(ctx)
}
