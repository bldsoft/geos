package source

import (
	"context"

	"github.com/bldsoft/geos/pkg/entity"
)

type Source interface {
	CheckUpdates(ctx context.Context) (entity.Updates, error)
	Download(ctx context.Context, update ...bool) (entity.Updates, error)
	DirPath() string
}
