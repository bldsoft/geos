package source2

import (
	"context"
	"fmt"
	"io"
	"path/filepath"

	"github.com/bldsoft/geos/pkg/entity"
	"github.com/bldsoft/geos/pkg/storage/maxmind/mmdb"
	"github.com/hashicorp/go-version"
)

const metadataChunkSize = 128 * 1024

type MMDBState struct {
	version    *version.Version
	buildEpoch uint
}

func (s MMDBState) IsHigher(other MMDBState) bool {
	if s.version.GreaterThan(other.version) {
		return true
	}

	return s.buildEpoch > other.buildEpoch
}

type MMDBSource struct {
	dbFile *UpdatableFile[MMDBState]
	name   entity.Subject
}

func NewMMDBSource(sourceUrl, dbPath string, name entity.Subject) *MMDBSource {
	res := &MMDBSource{
		dbFile: NewUpdatableFile(
			dbPath,
			sourceUrl,
			mmdbVersionFunc,
		),
		name: name,
	}

	return res
}

func (s *MMDBSource) Download(ctx context.Context) (entity.Updates, error) {
	err := s.dbFile.Update(ctx)
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

func (s *MMDBSource) CheckUpdates(ctx context.Context) (entity.Updates, error) {
	available, err := s.dbFile.CheckUpdates(ctx)
	return entity.Updates{s.name: &entity.UpdateStatus{Available: available, Error: err.Error()}}, nil
}

func (s *MMDBSource) DirPath() string {
	return filepath.Dir(s.dbFile.LocalPath)
}

func mmdbVersionFunc(ctx context.Context, path string, rep ReadFileRepository) (MMDBState, error) {
	r, err := rep.TailReader(ctx, path, metadataChunkSize)
	if err != nil {
		return MMDBState{}, err
	}
	defer r.Close()

	data, err := io.ReadAll(r)
	if err != nil {
		return MMDBState{}, err
	}

	return extractMMDBState(data)
}

func extractMMDBState(metadataBuf []byte) (MMDBState, error) {
	meta, err := mmdb.DecodeMetadata(metadataBuf)
	if err != nil {
		return MMDBState{}, fmt.Errorf("metadata decoding failed: %w", err)
	}

	v, err := version.NewVersion(fmt.Sprintf("%d.%d", meta.BinaryFormatMajorVersion, meta.BinaryFormatMinorVersion))
	if err != nil {
		return MMDBState{}, err
	}

	return MMDBState{
		version:    v,
		buildEpoch: meta.BuildEpoch,
	}, nil
}
