package geonames

import (
	"context"

	"github.com/bldsoft/geos/pkg/entity"
	"github.com/bldsoft/geos/pkg/storage/source"
)

type CustomStorage struct {
	*MultiStorage[*StoragePatch]
	source          source.Source
	archiveFilepath string
}

func NewCustomStorage(archiveFilepath string, patches ...*StoragePatch) *CustomStorage {
	return &CustomStorage{
		MultiStorage:    &MultiStorage[*StoragePatch]{storages: patches},
		archiveFilepath: archiveFilepath,
	}
}

func NewCustomStorageFromTarGz(archiveFilepath string) *CustomStorage {
	customs := NewStoragePatchesFromTarGz(archiveFilepath)
	return NewCustomStorage(archiveFilepath, customs...)
}

func (s *CustomStorage) SetSource(source source.Source) {
	s.source = source
}

func (s *CustomStorage) CheckUpdates(ctx context.Context) (entity.Updates, error) {
	if s.source == nil {
		return nil, source.ErrNoSource
	}

	return s.source.CheckUpdates(ctx)
}

func (s *CustomStorage) Download(ctx context.Context) (updates entity.Updates, err error) {
	if s.source == nil {
		return nil, source.ErrNoSource
	}

	if updates, err = s.source.Download(ctx); err != nil {
		return nil, err
	}

	s.storages = NewStoragePatchesFromTarGz(s.archiveFilepath)
	return updates, nil
}
