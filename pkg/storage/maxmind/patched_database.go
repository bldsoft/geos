package maxmind

import (
	"context"
	"errors"
	"net/netip"

	"github.com/bldsoft/geos/pkg/entity"
	"github.com/bldsoft/geos/pkg/utils"
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
	if custom != nil {
		db.Add(custom)
	}
	return db
}

func (db *PatchedDatabase) Lookup(ctx context.Context, ip netip.Addr, result any) error {
	err := db.MultiMaxMindDB.Lookup(ctx, ip, result)
	if errors.Is(err, utils.ErrNotFound) {
		return nil
	}
	return err
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

func (db *PatchedDatabase) CurrentVersion() entity.PatchedMMDBVersion {
	res := entity.PatchedMMDBVersion{DB: db.db.CurrentVersion()}
	if db.custom != nil {
		patch := entity.ModTimeVersion(db.custom.CurrentVersion())
		res.Patch = &patch
	}
	return res
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
		res.CurrentVersion.Patch = (*entity.ModTimeVersion)(&customUpdate.CurrentVersion)
		res.RemoteVersion.Patch = (*entity.ModTimeVersion)(&customUpdate.RemoteVersion)
	}

	return res, nil
}
