package geonames

import (
	"context"

	"github.com/bldsoft/geos/pkg/storage/maxmind"
)

type CustomStorage struct {
	*MultiStorage[*StoragePatch]
	source *StoragePatchesSource
}

func NewCustomStorage(patches ...*StoragePatch) *CustomStorage {
	return &CustomStorage{
		MultiStorage: &MultiStorage[*StoragePatch]{storages: patches},
	}
}

func NewCustomStorageFromDir(dir string) *CustomStorage {
	customs := NewStoragePatchesFromDir(dir, "geonames")
	return NewCustomStorage(customs...)
}

func (s *CustomStorage) SetSource(source *StoragePatchesSource) {
	s.source = source
}

func (s *CustomStorage) CheckUpdates(ctx context.Context) (bool, error) {
	if s.source == nil {
		return false, maxmind.ErrNoSource
	}

	return s.source.CheckUpdates(ctx)
}

func (s *CustomStorage) Download(ctx context.Context, update ...bool) error {
	if s.source == nil {
		return maxmind.ErrNoSource
	}

	if err := s.source.Download(ctx, update...); err != nil {
		return err
	}

	s.storages = NewStoragePatchesFromDir(s.source.storagePath, "geonames")
	return nil
}
