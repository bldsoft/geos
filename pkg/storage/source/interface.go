package source

import (
	"context"

	"github.com/bldsoft/geos/pkg/entity"
)

type Updater interface {
	Update(ctx context.Context, force bool) error
	CheckUpdates(ctx context.Context) (entity.Update, error)
}
