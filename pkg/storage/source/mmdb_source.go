package source

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/bldsoft/geos/pkg/storage/maxmind/mmdb"
	"github.com/hashicorp/go-version"
)

const metadataChunkSize = 128 * 1024

type MMDBVersion struct {
	version    *version.Version
	buildEpoch uint
}

func ParseMMDBVersion(s string) (MMDBVersion, error) {
	parts := strings.Split(s, "-")
	if len(parts) != 2 {
		return MMDBVersion{}, fmt.Errorf("invalid version format: %s", s)
	}

	version, err := version.NewVersion(parts[0])
	if err != nil {
		return MMDBVersion{}, err
	}

	buildEpoch, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		return MMDBVersion{}, err
	}

	return MMDBVersion{version: version, buildEpoch: uint(buildEpoch)}, nil
}

func (s MMDBVersion) IsHigher(other MMDBVersion) bool {
	if s.version.GreaterThan(other.version) {
		return true
	}

	return s.buildEpoch > other.buildEpoch
}

func (s MMDBVersion) String() string {
	return fmt.Sprintf("%s-%d", s.version.String(), s.buildEpoch)
}

type MMDBSource struct {
	dbFile *UpdatableFile[MMDBVersion]
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
		version:    v,
		buildEpoch: meta.BuildEpoch,
	}, nil
}
