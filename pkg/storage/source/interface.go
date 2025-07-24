package source

import (
	"context"

	"github.com/bldsoft/geos/pkg/entity"
)

type Updater interface {
	TryUpdate(ctx context.Context) error
	CheckUpdates(ctx context.Context) (entity.Update, error)
}

type RecoverableUpdater interface {
	Updater
	LastUpdateInterrupted(ctx context.Context) (bool, error)
}
