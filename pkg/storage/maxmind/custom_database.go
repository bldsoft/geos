package maxmind

import (
	"context"

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

func (db *CustomDatabase) Download(ctx context.Context, update ...bool) error {
	if db.source == nil {
		return ErrNoSource
	}

	if err := db.source.Download(ctx, update...); err != nil {
		return err
	}

	db.dbs = NewCustomDatabaseFromDir(db.source.DirPath(), db.customDBName).dbs
	return nil
}

func (db *CustomDatabase) CheckUpdates(ctx context.Context) (bool, error) {
	if db.source == nil {
		return false, ErrNoSource
	}
	return db.source.CheckUpdates(ctx)
}
