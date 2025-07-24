package source

import (
	"context"
	"io"
	"path/filepath"

	"github.com/bldsoft/geos/pkg/entity"
)

type PatchesSource struct {
	patchesFile *UpdatableFile[ModTimeVersion]
}

func NewPatchesSource(sourceUrl, dirPath, prefix string) *PatchesSource {
	return &PatchesSource{
		patchesFile: NewTSUpdatableFile(filepath.Join(dirPath, prefix+"_patches.tar.gz"), sourceUrl),
	}
}

func (s *PatchesSource) Name() string {
	return s.patchesFile.LocalPath
}

func (s *PatchesSource) Reader(ctx context.Context) (io.ReadCloser, error) {
	return s.patchesFile.Reader(ctx)
}

func (s *PatchesSource) Version(ctx context.Context) (ModTimeVersion, error) {
	return s.patchesFile.Version(ctx)
}

func (s *PatchesSource) CheckUpdates(ctx context.Context) (entity.Update, error) {
	return s.patchesFile.CheckUpdates(ctx)
}

func (s *PatchesSource) TryUpdate(ctx context.Context) error {
	return s.patchesFile.TryUpdate(ctx)
}

func (s *PatchesSource) LastUpdateInterrupted(ctx context.Context) (bool, error) {
	return s.patchesFile.LastUpdateInterrupted(ctx)
}
