package maxmind

import (
	"context"

	"github.com/bldsoft/geos/pkg/entity"
	"github.com/bldsoft/gost/log"
)

type PatchedDatabase struct {
	*MultiMaxMindDB
	db     *MaxmindDatabase
	custom *CustomDatabase
}

func NewPatchedDatabase(db *MaxmindDatabase) *PatchedDatabase {
	return &PatchedDatabase{
		MultiMaxMindDB: NewMultiMaxMindDB(db),
		db:             db,
	}
}

func (db *PatchedDatabase) SetCustom(custom *CustomDatabase) *PatchedDatabase {
	db.custom = custom
	return db
}

func (db *PatchedDatabase) WithLogger(logger log.ServiceLogger) *PatchedDatabase {
	db.MultiMaxMindDB.WithLogger(logger)
	return db
}

func (db *PatchedDatabase) Update(ctx context.Context, force bool) error {
	if err := db.db.Update(ctx, force); err != nil {
		return err
	}

	if db.custom != nil {
		return db.custom.Update(ctx, force)
	}
	return nil
}

func (db *PatchedDatabase) CheckUpdates(ctx context.Context) (entity.Update[entity.PatchedMMDBVersion], error) {
	dbUpdate, err := db.db.CheckUpdates(ctx)
	if err != nil {
		return entity.Update[entity.PatchedMMDBVersion]{}, err
	}
	res := entity.Update[entity.PatchedMMDBVersion]{
		CurrentVersion: entity.PatchedMMDBVersion{DB: dbUpdate.CurrentVersion},
		RemoteVersion:  entity.PatchedMMDBVersion{DB: dbUpdate.RemoteVersion},
	}

	if db.custom != nil {
		customUpdate, err := db.custom.CheckUpdates(ctx)
		if err != nil {
			return entity.Update[entity.PatchedMMDBVersion]{}, err
		}
		res.CurrentVersion.Patch = entity.ModTimeVersion(customUpdate.CurrentVersion)
		res.RemoteVersion.Patch = entity.ModTimeVersion(customUpdate.RemoteVersion)
	}

	return res, nil
}
