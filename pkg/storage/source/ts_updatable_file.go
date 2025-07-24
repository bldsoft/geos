package source

import (
	"context"
	"fmt"
	"time"
)

func NewTSUpdatableFile(path, url string) *UpdatableFile[ModTimeVersion] {
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

func (v ModTimeVersion) IsHigher(other ModTimeVersion) bool {
	return time.Time(v).After(time.Time(other))
}

func (v ModTimeVersion) Time() time.Time {
	return time.Time(v)
}

func (v ModTimeVersion) String() string {
	if v.Time().IsZero() {
		return "0"
	}
	return fmt.Sprintf("%d", v.Time().Unix())
}
