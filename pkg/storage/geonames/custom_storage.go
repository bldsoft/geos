package geonames

import (
	"context"
	"os"

	"github.com/bldsoft/geos/pkg/entity"
	"github.com/bldsoft/geos/pkg/storage/source"
	"github.com/bldsoft/geos/pkg/storage/state"
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

func (s *CustomStorage) State() *state.GeosState {
	result := &state.GeosState{}

	var archiveTimestamp int64
	if info, err := os.Stat(s.archiveFilepath); err == nil {
		archiveTimestamp = info.ModTime().Unix()
	} else {
		for _, storage := range s.storages {
			storageState := storage.State()
			if storageState.GeonamesPatchesTimestamp > 0 {
				archiveTimestamp = storageState.GeonamesPatchesTimestamp
				break
			}
		}
	}

	if archiveTimestamp > 0 {
		result.GeonamesPatchesTimestamp = archiveTimestamp
	}

	return result
}
