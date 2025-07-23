package source

import (
	"context"

	"github.com/bldsoft/geos/pkg/entity"
	"github.com/bldsoft/geos/pkg/storage/state"
)

type Updater interface {
	Download(ctx context.Context) (entity.Updates, error)
	CheckUpdates(ctx context.Context) (entity.Updates, error)
}

type RecoverableUpdater interface {
	Updater
	LastUpdateInterrupted(ctx context.Context) (bool, error)
}

type Stater interface { // TODO: move??
	State() *state.GeosState
}
