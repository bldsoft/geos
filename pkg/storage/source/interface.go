package source

import (
	"context"
	"fmt"

	"github.com/bldsoft/geos/pkg/entity"
	"github.com/bldsoft/geos/pkg/storage/state"
)

var ErrNoSource = fmt.Errorf("no source provided for the database")

type Updater interface {
	Download(ctx context.Context) (entity.Updates, error)
	CheckUpdates(ctx context.Context) (entity.Updates, error)
}

type Source interface {
	Updater
	DirPath() string
}

type Stater interface {
	State() *state.GeosState
}
