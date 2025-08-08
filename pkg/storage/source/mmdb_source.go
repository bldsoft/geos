package source

import (
	"cmp"
	"context"
	"fmt"
	"io"

	"github.com/bldsoft/geos/pkg/storage/maxmind/mmdb"
	"github.com/hashicorp/go-version"
)

const metadataChunkSize = 128 * 1024

type MMDBVersion struct {
	Version    *version.Version
	BuildEpoch uint
}

func (s MMDBVersion) Compare(other MMDBVersion) int {
	if s.Version == nil {
		return -1
	}
	if other.Version == nil {
		return 1
	}
	return cmp.Or(
		s.Version.Compare(other.Version),
		cmp.Compare(s.BuildEpoch, other.BuildEpoch),
	)
}

type MMDBSource struct {
	dbFile *UpdatableFile[MMDBVersion]
}

func NewMMDBSource(dbPath, sourceUrl string) *MMDBSource {
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

func (s *MMDBSource) Version(ctx context.Context) (MMDBVersion, error) {
	return s.dbFile.Version(ctx)
}

func (s *MMDBSource) Update(ctx context.Context, force bool) error {
	return s.dbFile.Update(ctx, force)
}

func (s *MMDBSource) CheckUpdates(ctx context.Context) (Update[MMDBVersion], error) {
	return s.dbFile.CheckUpdates(ctx)
}

func mmdbVersionFunc(ctx context.Context, path string, rep ReadFileRepository) (MMDBVersion, error) {
	r, err := rep.TailReader(ctx, path, metadataChunkSize)
	if err != nil {
		return MMDBVersion{}, err
	}
	defer r.Close()

	data, err := io.ReadAll(r)
	if err != nil {
		return MMDBVersion{}, err
	}

	return extractMMDBState(data)
}

func extractMMDBState(metadataBuf []byte) (MMDBVersion, error) {
	meta, err := mmdb.DecodeMetadata(metadataBuf)
	if err != nil {
		return MMDBVersion{}, fmt.Errorf("metadata decoding failed: %w", err)
	}

	v, err := version.NewVersion(fmt.Sprintf("%d.%d", meta.BinaryFormatMajorVersion, meta.BinaryFormatMinorVersion))
	if err != nil {
		return MMDBVersion{}, err
	}

	return MMDBVersion{
		Version:    v,
		BuildEpoch: meta.BuildEpoch,
	}, nil
}
