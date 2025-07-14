package source2

import (
	"context"
	"path/filepath"

	"github.com/bldsoft/geos/pkg/entity"
)

type PatchesSource struct {
	patchesFile *UpdatableFile[ModTimeVersion]
	name        entity.Subject
	prefix      string
}

func NewPatchesSource(sourceUrl, dirPath, prefix string, name entity.Subject) *PatchesSource {
	// TODO: Recover interrupted downloads
	return &PatchesSource{
		patchesFile: NewTSUpdatableFile(filepath.Join(dirPath, prefix+"_patches.tar.gz"), sourceUrl),
		name:        name,
		prefix:      prefix,
	}
}

func (s *PatchesSource) Download(ctx context.Context) (entity.Updates, error) {
	err := s.patchesFile.Update(ctx)
	if err != nil {
		return entity.Updates{
			s.name: &entity.UpdateStatus{
				Error: err.Error(),
			},
		}, nil
	}
	return entity.Updates{
		s.name: &entity.UpdateStatus{},
	}, nil
}

func (s *PatchesSource) CheckUpdates(ctx context.Context) (entity.Updates, error) {
	available, err := s.patchesFile.CheckUpdates(ctx)
	return entity.Updates{s.name: &entity.UpdateStatus{Available: available, Error: err.Error()}}, nil
}

func (s *PatchesSource) DirPath() string {
	return filepath.Dir(s.patchesFile.LocalPath)
}
