package source

import (
	"context"
	"fmt"
	"io"

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

func (s MMDBState) String() string {
	return fmt.Sprintf("%s-%d", s.version.String(), s.buildEpoch)
}

type MMDBSource struct {
	dbFile *UpdatableFile[MMDBState]
}

func NewMMDBSource(sourceUrl, dbPath string) *MMDBSource {
	res := &MMDBSource{
		dbFile: NewUpdatableFile(
			dbPath,
			sourceUrl,
			mmdbVersionFunc,
		),
	}

	return res
}

func (s *MMDBSource) Reader(ctx context.Context) (io.ReadCloser, error) {
	return s.dbFile.Reader(ctx)
}

func (s *MMDBSource) LastUpdateInterrupted(ctx context.Context) (bool, error) {
	return s.dbFile.LastUpdateInterrupted(ctx)
}

func (s *MMDBSource) TryUpdate(ctx context.Context) error {
	return s.dbFile.TryUpdate(ctx)
}

func (s *MMDBSource) CheckUpdates(ctx context.Context) (entity.Update, error) {
	return s.dbFile.CheckUpdates(ctx)
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
