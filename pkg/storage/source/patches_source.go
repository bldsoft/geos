package source

import (
	"context"
	"io"
	"path/filepath"

	"github.com/bldsoft/geos/pkg/entity"
)

type PatchesSource struct {
	Name        entity.Subject
	patchesFile *UpdatableFile[ModTimeVersion]
}

func NewPatchesSource(sourceUrl, dirPath, prefix string, name entity.Subject) *PatchesSource {
	return &PatchesSource{
		patchesFile: NewTSUpdatableFile(filepath.Join(dirPath, prefix+"_patches.tar.gz"), sourceUrl),
		Name:        name,
	}
}

func (s *PatchesSource) Reader(ctx context.Context) (io.ReadCloser, error) {
	return s.patchesFile.Reader(ctx)
}

func (s *PatchesSource) Version(ctx context.Context) (ModTimeVersion, error) {
	return s.patchesFile.Version(ctx)
}

func (s *PatchesSource) LastUpdateInterrupted(ctx context.Context) (bool, error) {
	return s.patchesFile.LastUpdateInterrupted(ctx)
}

func (s *PatchesSource) Download(ctx context.Context) (entity.Updates, error) {
	upd := entity.Updates{}
	updated, err := s.patchesFile.Update(ctx)
	if err != nil {
		upd[s.Name] = &entity.UpdateStatus{Error: err.Error()}
		return upd, nil
	}
	if updated {
		upd[s.Name] = &entity.UpdateStatus{}
	}
	return upd, nil
}

func (s *PatchesSource) CheckUpdates(ctx context.Context) (entity.Updates, error) {
	upd := entity.Updates{}
	available, err := s.patchesFile.CheckUpdates(ctx)
	if err != nil {
		upd[s.Name] = &entity.UpdateStatus{Error: err.Error()}
		return upd, nil
	}
	if available {
		upd[s.Name] = &entity.UpdateStatus{Available: true}
	}
	return upd, nil
}
