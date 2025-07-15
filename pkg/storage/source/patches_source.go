package source

import (
	"context"
	"path/filepath"

	"github.com/bldsoft/geos/pkg/entity"
)

type PatchesSource struct {
	Name        entity.Subject
	patchesFile *UpdatableFile[ModTimeVersion]
	prefix      string
}

func NewPatchesSource(sourceUrl, dirPath, prefix string, name entity.Subject) *PatchesSource {
	return &PatchesSource{
		patchesFile: NewTSUpdatableFile(filepath.Join(dirPath, prefix+"_patches.tar.gz"), sourceUrl),
		Name:        name,
		prefix:      prefix,
	}
}

func (s *PatchesSource) HasBeenInterrupted() bool {
	return s.patchesFile.HasBeenInterrupted()
}

func (s *PatchesSource) ArchiveFilePath() string {
	return s.patchesFile.LocalPath
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
