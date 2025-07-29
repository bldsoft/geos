package source

import (
	"context"
	"fmt"
	"strconv"
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

func ParseModTimeVersion(s string) (ModTimeVersion, error) {
	unix, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return ModTimeVersion{}, err
	}
	return ModTimeVersion(time.Unix(unix, 0)), nil
}

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
