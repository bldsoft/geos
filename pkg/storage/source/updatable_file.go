package source

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"
)

var ErrFileExists = errors.New("file exists")

type Comparable[T any] interface {
	IsHigher(other T) bool
}

type ReadFileRepository interface {
	Reader(ctx context.Context, path string) (io.ReadCloser, error)
	TailReader(ctx context.Context, path string, offset int64) (io.ReadCloser, error)
	LastModified(ctx context.Context, path string) (time.Time, error)
	Exists(ctx context.Context, path string) (bool, error)
}

type WriteFileRepository interface {
	Open(ctx context.Context, path string) (io.WriteCloser, error)
	Write(ctx context.Context, path string, reader io.Reader) error
	Rename(ctx context.Context, oldPath, newPath string) error
	Remove(ctx context.Context, path string) error
}

type FileRepository interface {
	ReadFileRepository
	WriteFileRepository
}

type UpdatableFile[V Comparable[V]] struct {
	LocalPath string
	RemoteURL string

	LocalFileRepository  FileRepository
	RemoteFileRepository ReadFileRepository

	VersionFunc func(ctx context.Context, path string, rep ReadFileRepository) (V, error)
}

func NewUpdatableFile[V Comparable[V]](
	path,
	url string,
	versionFunc func(ctx context.Context, path string, rep ReadFileRepository) (V, error),
) *UpdatableFile[V] {
	return &UpdatableFile[V]{
		LocalPath:            path,
		RemoteURL:            url,
		LocalFileRepository:  NewLocalFileRepository(),
		RemoteFileRepository: NewRemoteFileRepository(),
		VersionFunc:          versionFunc,
	}
}

func (u *UpdatableFile[V]) Reader(ctx context.Context) (io.ReadCloser, error) {
	return u.LocalFileRepository.Reader(ctx, u.LocalPath)
}

func (u *UpdatableFile[V]) Version(ctx context.Context) (V, error) {
	return u.VersionFunc(ctx, u.LocalPath, u.LocalFileRepository)
}

func (u *UpdatableFile[V]) RemoteVersion(ctx context.Context) (V, error) {
	return u.VersionFunc(ctx, u.RemoteURL, u.RemoteFileRepository)
}

func (u *UpdatableFile[V]) Update(ctx context.Context) (bool, error) {
	if need, err := u.needUpdate(ctx); !need || err != nil {
		return false, err
	}
	err := u.update(ctx)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (u *UpdatableFile[V]) update(ctx context.Context) error {
	if err := u.downloadTempFile(ctx); err != nil {
		return err
	}
	defer u.LocalFileRepository.Remove(ctx, u.tmpFilePath())

	if err := u.LocalFileRepository.Rename(ctx, u.tmpFilePath(), u.LocalPath); err != nil {
		return fmt.Errorf("failed to move temporary file: %w", err)
	}

	return nil
}

func (u *UpdatableFile[V]) downloadTempFile(ctx context.Context) error {
	tmpFile, err := u.LocalFileRepository.Open(ctx, u.tmpFilePath())
	if err != nil {
		if errors.Is(err, ErrFileExists) {
			return nil
		}
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer tmpFile.Close()

	reader, err := u.RemoteFileRepository.Reader(ctx, u.RemoteURL)
	if err != nil {
		return err
	}
	defer reader.Close()

	if _, err := io.Copy(tmpFile, reader); err != nil {
		os.Remove(u.tmpFilePath())
		return fmt.Errorf("failed to download file: %w", err)
	}

	return nil
}

func (u *UpdatableFile[V]) needUpdate(ctx context.Context) (bool, error) {
	remoteVersion, err := u.RemoteVersion(ctx)
	if err != nil {
		return false, err
	}

	localVersion, err := u.Version(ctx)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return true, nil
		}
		return false, err
	}

	return remoteVersion.IsHigher(localVersion), nil
}

func (u *UpdatableFile[V]) tmpFilePath() string {
	return u.LocalPath + ".tmp"
}

func (u *UpdatableFile[V]) CheckUpdates(ctx context.Context) (available bool, err error) {
	remoteVersion, err := u.RemoteVersion(ctx)
	if err != nil {
		return false, err
	}

	localVersion, err := u.Version(ctx)
	if err != nil {
		return false, err
	}

	return remoteVersion.IsHigher(localVersion), nil
}

func (u *UpdatableFile[V]) HasBeenInterrupted() bool {
	exists, _ := u.LocalFileRepository.Exists(context.Background(), u.tmpFilePath())
	return exists
}
