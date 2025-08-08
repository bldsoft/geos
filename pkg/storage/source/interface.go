package source

import (
	"context"
	"io"
	"time"
)

type Comparable[T any] interface {
	Compare(other T) int
}

type Update[V Comparable[V]] struct {
	CurrentVersion V
	RemoteVersion  V
}

type Updater[V Comparable[V]] interface {
	Update(ctx context.Context, force bool) error
	CheckUpdates(ctx context.Context) (Update[V], error)
}

type ReadFileRepository interface {
	Reader(ctx context.Context, path string) (io.ReadCloser, error)
	TailReader(ctx context.Context, path string, offset int64) (io.ReadCloser, error)
	LastModified(ctx context.Context, path string) (time.Time, error)
	Exists(ctx context.Context, path string) (bool, error)
}

type WriteFileRepository interface {
	CreateIfNotExists(ctx context.Context, path string) (io.WriteCloser, error)
	Write(ctx context.Context, path string, reader io.Reader) error
	Rename(ctx context.Context, oldPath, newPath string) error
	Remove(ctx context.Context, path string) error
}

type FileRepository interface {
	ReadFileRepository
	WriteFileRepository
}
