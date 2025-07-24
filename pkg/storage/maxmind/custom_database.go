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
	source *source.PatchesSource
}

func NewCustomDatabaseFromTarGz(source *source.PatchesSource) *CustomDatabase {
	customDBs, err := NewDatabasePatchesFromTarGz(source)
	if err != nil {
		log.Logger.Errorf("failed to load custom databases %s: %v", source.Name(), err)
	}
	res := &CustomDatabase{
		MultiMaxMindDB: NewMultiMaxMindDB(customDBs...),
		source:         source,
	}
	return res
}

func (db *CustomDatabase) MetaData() (*maxminddb.Metadata, error) {
	if len(db.dbs) == 0 {
		return &maxminddb.Metadata{}, nil //there is no problem if there are no patches
	}
	return db.MultiMaxMindDB.MetaData()
}

func (db *CustomDatabase) TryUpdate(ctx context.Context) error {
	if err := db.source.TryUpdate(ctx); err != nil {
		return err
	}

	customDBs, err := NewDatabasePatchesFromTarGz(db.source)
	if err != nil {
		return err
	}

	db.MultiMaxMindDB = NewMultiMaxMindDB(customDBs...)
	return nil
}

func (db *CustomDatabase) CheckUpdates(ctx context.Context) (entity.Update, error) {
	return db.source.CheckUpdates(ctx)
}

func (db *CustomDatabase) LastUpdateInterrupted(ctx context.Context) (bool, error) {
	return db.source.LastUpdateInterrupted(ctx)
}
