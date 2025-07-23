package geonames

import (
	"context"

	"github.com/bldsoft/geos/pkg/entity"
	"github.com/bldsoft/geos/pkg/storage/source"
	"github.com/bldsoft/geos/pkg/storage/state"
)

type CustomStorage struct {
	*MultiStorage[*StoragePatch]
	source *source.PatchesSource
}

func NewCustomStorageFromTarGz(source *source.PatchesSource) *CustomStorage {
	patches := NewStoragePatchesFromTarGz(source)
	return &CustomStorage{
		MultiStorage: NewMultiStorage[*StoragePatch](patches...),
		source:       source,
	}
}

func (s *CustomStorage) CheckUpdates(ctx context.Context) (entity.Updates, error) {
	return s.source.CheckUpdates(ctx)
}

func (s *CustomStorage) Download(ctx context.Context) (updates entity.Updates, err error) {
	if updates, err = s.source.Download(ctx); err != nil {
		return nil, err
	}

	patches := NewStoragePatchesFromTarGz(s.source)
	s.MultiStorage = NewMultiStorage(patches...)
	return updates, nil
}

func (s *CustomStorage) State() *state.GeosState {
	ctx := context.Background()
	result := &state.GeosState{}

	var archiveTimestamp int64
	if v, err := s.source.Version(ctx); err == nil {
		archiveTimestamp = v.Time().Unix()
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
