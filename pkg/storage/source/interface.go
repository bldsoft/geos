package source

import "context"

type Source interface {
	CheckUpdates(ctx context.Context) (bool, error)
	Download(ctx context.Context, update ...bool) error
	DirPath() string
}
