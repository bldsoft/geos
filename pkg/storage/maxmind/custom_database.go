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
	source          *source.PatchesSource
	archiveFilepath string
}

func NewCustomDatabase(archiveFilepath string, patches ...*DatabasePatch) *CustomDatabase {
	return &CustomDatabase{
		archiveFilepath: archiveFilepath,
		MultiMaxMindDB:  &MultiMaxMindDB[*DatabasePatch]{dbs: patches, logger: log.Logger},
	}
}

func NewCustomDatabaseFromTarGz(archiveFilepath string) *CustomDatabase {
	customDBs, err := NewDatabasePatchesFromTarGz(archiveFilepath)
	if err != nil {
		log.Logger.Errorf("failed to load custom databases from %s: %v", archiveFilepath, err)
	}

	return NewCustomDatabase(archiveFilepath, customDBs...)
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

func (db *CustomDatabase) Download(ctx context.Context) (updates entity.Updates, err error) {
	if db.source == nil {
		return nil, source.ErrNoSource
	}

	if updates, err = db.source.Download(ctx); err != nil {
		return nil, err
	}

	db.dbs = NewCustomDatabaseFromTarGz(db.archiveFilepath).dbs
	return
}

func (db *CustomDatabase) CheckUpdates(ctx context.Context) (entity.Updates, error) {
	if db.source == nil {
		return nil, source.ErrNoSource
	}
	return db.source.CheckUpdates(ctx)
}
