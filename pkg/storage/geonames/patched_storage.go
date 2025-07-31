package geonames

import (
	"context"

	"github.com/bldsoft/geos/pkg/entity"
)

type PatchedStorage struct {
	MultiStorage
	storage *GeoNameStorage
	custom  *CustomStorage
}

func NewPatchedStorage(storage *GeoNameStorage) *PatchedStorage {
	return &PatchedStorage{
		MultiStorage: *NewMultiStorage(storage),
		storage:      storage,
	}
}

func (s *PatchedStorage) Add(custom *CustomStorage) *PatchedStorage {
	s.custom = custom
	return s
}

func (s *PatchedStorage) CheckUpdates(ctx context.Context) (entity.Update[entity.PatchedGeoNamesVersion], error) {
	dbUpdate, err := s.storage.CheckUpdates(ctx)
	if err != nil {
		return entity.Update[entity.PatchedGeoNamesVersion]{}, err
	}
	res := entity.Update[entity.PatchedGeoNamesVersion]{
		CurrentVersion: entity.PatchedGeoNamesVersion{
			DB: entity.GeoNamesVersion(dbUpdate.CurrentVersion),
		},
		RemoteVersion: entity.PatchedGeoNamesVersion{
			DB: entity.GeoNamesVersion(dbUpdate.RemoteVersion),
		},
	}
	if s.custom != nil {
		customUpdate, err := s.custom.CheckUpdates(ctx)
		if err != nil {
			return entity.Update[entity.PatchedGeoNamesVersion]{}, err
		}
		res.CurrentVersion.PatchVersion = entity.ModTimeVersion(customUpdate.CurrentVersion)
		res.RemoteVersion.PatchVersion = entity.ModTimeVersion(customUpdate.RemoteVersion)
	}
	return res, nil
}

func (s *PatchedStorage) Update(ctx context.Context, force bool) error {
	if err := s.storage.Update(ctx, force); err != nil {
		return err
	}
	if s.custom != nil {
		return s.custom.Update(ctx, force)
	}
	return nil
}
