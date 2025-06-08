package source

import (
	"context"
	"fmt"

	"github.com/bldsoft/geos/pkg/entity"
)

var ErrNoSource = fmt.Errorf("no source provided for the database")

type Source interface {
	CheckUpdates(ctx context.Context) (entity.Updates, error)
	Download(ctx context.Context, update ...bool) (entity.Updates, error)
	DirPath() string
}
