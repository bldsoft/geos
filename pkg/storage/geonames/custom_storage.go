package geonames

import (
	"context"

	"github.com/bldsoft/geos/pkg/entity"
	"github.com/bldsoft/geos/pkg/storage/source"
)

type CustomStorage struct {
	*MultiStorage[*StoragePatch]
	source *source.PatchesSource
}

func NewCustomStorageFromTarGz(source *source.PatchesSource) *CustomStorage {
	patches := NewStoragePatchesFromTarGz(source)
	return &CustomStorage{
		MultiStorage: NewMultiStorage(patches...),
		source:       source,
	}
}

func (s *CustomStorage) CheckUpdates(ctx context.Context) (entity.Update, error) {
	return s.source.CheckUpdates(ctx)
}

func (s *CustomStorage) Update(ctx context.Context, force bool) error {
	update, err := s.CheckUpdates(ctx)
	if err != nil {
		return err
	}
	if update.AvailableVersion == "" {
		return nil
	}

	if err := s.source.Update(ctx, force); err != nil {
		return err
	}

	patches := NewStoragePatchesFromTarGz(s.source)
	s.MultiStorage = NewMultiStorage(patches...)
	return nil
}
