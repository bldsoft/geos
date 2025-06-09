package source

import (
	"context"
	"fmt"

	"github.com/bldsoft/geos/pkg/entity"
)

var ErrNoSource = fmt.Errorf("no source provided for the database")

type Updater interface {
	Download(ctx context.Context, update ...bool) (entity.Updates, error)
	CheckUpdates(ctx context.Context) (entity.Updates, error)
}

type Source interface {
	Updater
	DirPath() string
}
