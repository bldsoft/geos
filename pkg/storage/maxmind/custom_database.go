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
	source     *source.TSUpdatableFile
	lastUpdate source.ModTimeVersion
}

func NewCustomDatabaseFromTarGz(source *source.TSUpdatableFile) *CustomDatabase {
	logger := log.Logger.WithFields(log.Fields{"source": source.LocalPath, "db": "custom maxmind"})

	version, err := source.Version(context.Background())
	if err != nil {
		logger.Errorf("failed to get local version: %v", err)
	}
	customDBs, err := NewDatabasePatchesFromTarGz(source)
	if err != nil {
		logger.Errorf("failed to get patches: %v", err)
	}
	res := &CustomDatabase{
		MultiMaxMindDB: NewMultiMaxMindDB(customDBs...),
		source:         source,
		lastUpdate:     version,
	}
	return res
}

func (db *CustomDatabase) MetaData() (*maxminddb.Metadata, error) {
	if len(db.dbs) == 0 {
		return &maxminddb.Metadata{}, nil //there is no problem if there are no patches
	}
	return db.MultiMaxMindDB.MetaData()
}

func (db *CustomDatabase) Update(ctx context.Context, force bool) error {
	update, err := db.CheckUpdates(ctx)
	if err != nil {
		return err
	}
	if update.RemoteVersion != "" {
		return nil
	}

	if err := db.source.Update(ctx, force); err != nil {
		return err
	}

	version, err := db.source.Version(ctx)
	if err != nil {
		return err
	}

	customDBs, err := NewDatabasePatchesFromTarGz(db.source)
	if err != nil {
		return err
	}

	db.MultiMaxMindDB = NewMultiMaxMindDB(customDBs...)
	db.lastUpdate = version
	return nil
}

func (db *CustomDatabase) CheckUpdates(ctx context.Context) (entity.Update, error) {
	update, err := db.source.CheckUpdates(ctx)
	if err != nil {
		return entity.Update{}, err
	}
	if update.RemoteVersion != "" {
		return update, nil
	}

	version, err := db.source.Version(ctx)
	if err != nil {
		return entity.Update{}, err
	}

	if !version.IsHigher(db.lastUpdate) {
		return update, nil
	}

	return entity.Update{
		CurrentVersion: db.lastUpdate.String(),
		RemoteVersion:  version.String(),
	}, nil
}
