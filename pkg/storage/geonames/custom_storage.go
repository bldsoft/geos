package geonames

import (
	"context"
	"path/filepath"
	"sync/atomic"

	"github.com/bldsoft/geos/pkg/entity"
	"github.com/bldsoft/geos/pkg/storage/source"
	"github.com/bldsoft/gost/log"
)

type CustomStorage struct {
	base       atomic.Pointer[MultiStorage]
	source     *source.TSUpdatableFile
	lastUpdate source.ModTimeVersion
}

func NewCustomStorage(ctx context.Context, source *source.TSUpdatableFile) *CustomStorage {
	ctx = context.WithValue(ctx, log.LoggerCtxKey, log.FromContext(ctx).WithFields(log.Fields{"type": "patch"}))

	res := &CustomStorage{
		source: source,
	}
	res.base.Store(NewMultiStorage())

	if err := res.update(ctx); err != nil {
		log.FromContext(ctx).Errorf("failed to get patches: %v", err)
	}
	return res
}

func (s *CustomStorage) CheckUpdates(ctx context.Context) (entity.Update[source.ModTimeVersion], error) {
	update, err := s.source.CheckUpdates(ctx)
	if err != nil {
		return entity.Update[source.ModTimeVersion]{}, err
	}
	update.CurrentVersion = s.lastUpdate
	return update, nil
}

func (s *CustomStorage) Update(ctx context.Context, force bool) error {
	update, err := s.source.CheckUpdates(ctx)
	if err != nil {
		return err
	}

	if update.RemoteVersion.Compare(update.CurrentVersion) > 0 {
		if err := s.source.Update(ctx, force); err != nil {
			return err
		}
	}

	if update.RemoteVersion.Compare(s.lastUpdate) > 0 {
		if err := s.update(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (s *CustomStorage) update(ctx context.Context) error {
	var patches []Storage
	var err error
	if filepath.Ext(s.source.LocalPath) == ".json" {
		patch, err := NewStoragePatchFromJSON(s.source)
		if err != nil {
			return err
		}
		patches = []Storage{patch}
	} else {
		patches, err = NewStoragePatchesFromTarGz(s.source)
	}
	if err != nil {
		return err
	}

	version, err := s.source.Version(ctx)
	if err != nil {
		return err
	}
	s.base.Store(NewMultiStorage(patches...))
	s.lastUpdate = version
	return nil
}
