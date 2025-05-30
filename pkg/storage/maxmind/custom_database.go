package maxmind

import (
	"context"

	"github.com/bldsoft/gost/log"
	"github.com/oschwald/maxminddb-golang"
)

// stores patches for easier updates
type CustomDatabase struct {
	*MultiMaxMindDB[*DatabasePatch]
	source *DBPatchesSource
}

func NewCustomDatabase(patches ...*DatabasePatch) *CustomDatabase {
	return &CustomDatabase{
		MultiMaxMindDB: &MultiMaxMindDB[*DatabasePatch]{dbs: patches, logger: log.Logger},
	}
}

func NewCustomDatabaseFromDir(dir, customDBPrefix string) *CustomDatabase {
	customDBs := NewDatabasePatchesFromDir(dir, customDBPrefix)
	return NewCustomDatabase(customDBs...)
}

func (db *CustomDatabase) SetSource(source *DBPatchesSource) {
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

	return db.source.Download(ctx, update...)
}

func (db *CustomDatabase) CheckUpdates(ctx context.Context) (bool, error) {
	if db.source == nil {
		return false, ErrNoSource
	}
	return db.source.CheckUpdates(ctx)
}
