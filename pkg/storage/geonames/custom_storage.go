package geonames

import (
	"context"

	"github.com/bldsoft/geos/pkg/entity"
	"github.com/bldsoft/geos/pkg/storage/source"
	"github.com/bldsoft/gost/log"
)

type CustomStorage struct {
	*MultiStorage[*StoragePatch]
	source     *source.TSUpdatableFile
	lastUpdate source.ModTimeVersion
}

func NewCustomStorageFromTarGz(source *source.TSUpdatableFile) *CustomStorage {
	logger := log.Logger.WithFields(log.Fields{"source": source.LocalPath, "db": "custom geonames"})

	version, err := source.Version(context.Background())
	if err != nil {
		logger.Errorf("failed to get local version: %v", err)
	}

	patches, err := NewStoragePatchesFromTarGz(source)
	if err != nil {
		logger.Errorf("failed to get patches: %v", err)
	}

	return &CustomStorage{
		MultiStorage: NewMultiStorage(patches...),
		source:       source,
		lastUpdate:   version,
	}
}

func (s *CustomStorage) CheckUpdates(ctx context.Context) (entity.Update, error) {
	update, err := s.source.CheckUpdates(ctx)
	if err != nil {
		return entity.Update{}, err
	}
	if update.RemoteVersion != "" {
		return update, nil
	}

	version, err := s.source.Version(ctx)
	if err != nil {
		return entity.Update{}, err
	}

	if !version.IsHigher(s.lastUpdate) {
		return update, nil
	}

	return entity.Update{
		CurrentVersion: s.lastUpdate.String(),
		RemoteVersion:  version.String(),
	}, nil
}

func (s *CustomStorage) Update(ctx context.Context, force bool) error {
	update, err := s.source.CheckUpdates(ctx)
	if err != nil {
		return err
	}

	if update.RemoteVersion != "" {
		if err := s.source.Update(ctx, force); err != nil {
			return err
		}
	}

	version, err := s.source.Version(ctx)
	if err != nil {
		return err
	}

	if !version.IsHigher(s.lastUpdate) {
		return nil
	}

	patches, err := NewStoragePatchesFromTarGz(s.source)
	if err != nil {
		return err
	}
	s.MultiStorage = NewMultiStorage(patches...)
	s.lastUpdate = version
	return nil
}
