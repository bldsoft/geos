package source

import (
	"context"
	"time"
)

type TSUpdatableFile = UpdatableFile[ModTimeVersion]

func NewTSUpdatableFile(path, url string) *TSUpdatableFile {
	return NewUpdatableFile(
		path,
		url,
		func(ctx context.Context, path string, rep ReadFileRepository) (ModTimeVersion, error) {
			r, err := rep.LastModified(ctx, path)
			if err != nil {
				return ModTimeVersion{}, err
			}
			return ModTimeVersion(r), nil
		},
	)
}

type ModTimeVersion time.Time

func (v ModTimeVersion) Compare(other ModTimeVersion) int {
	return time.Time(v).Compare(time.Time(other))
}

func (v ModTimeVersion) Time() time.Time {
	return time.Time(v)
}
