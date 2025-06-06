package geonames

import (
	"context"

	"github.com/bldsoft/geos/pkg/entity"
	"github.com/bldsoft/geos/pkg/storage/maxmind"
	"github.com/bldsoft/geos/pkg/storage/source"
)

type CustomStorage struct {
	*MultiStorage[*StoragePatch]
	source source.Source
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

func (s *CustomStorage) SetSource(source source.Source) {
	s.source = source
}

func (s *CustomStorage) CheckUpdates(ctx context.Context) (entity.Updates, error) {
	if s.source == nil {
		return nil, maxmind.ErrNoSource
	}

	return s.source.CheckUpdates(ctx)
}

func (s *CustomStorage) Download(ctx context.Context, update ...bool) (updates entity.Updates, err error) {
	if s.source == nil {
		return nil, maxmind.ErrNoSource
	}

	if updates, err = s.source.Download(ctx, update...); err != nil {
		return nil, err
	}

	s.storages = NewStoragePatchesFromDir(s.source.DirPath(), "geonames")
	return updates, nil
}
