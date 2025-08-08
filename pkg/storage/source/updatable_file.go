package source

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/bldsoft/geos/pkg/utils"
	"github.com/bldsoft/gost/log"
)

var ErrFileExists = errors.New("file exists")
var ErrFileNotExists = errors.New("file not exists")
var ErrRemoteURLNotSet = errors.New("remote URL is not set")

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
	reader, err := u.LocalFileRepository.Reader(ctx, u.LocalPath)
	if err == nil {
		return reader, nil
	}

	if !errors.Is(err, ErrFileNotExists) {
		return nil, fmt.Errorf("failed to open local file: %w", err)
	}

	log.FromContext(ctx).Info("Local file missing, starting download")

	if err := u.update(ctx, true); err != nil {
		return nil, fmt.Errorf("failed to update file: %w", err)
	}

	return u.LocalFileRepository.Reader(ctx, u.LocalPath)
}

func (u *UpdatableFile[V]) Version(ctx context.Context) (V, error) {
	return u.VersionFunc(ctx, u.LocalPath, u.LocalFileRepository)
}

func (u *UpdatableFile[V]) RemoteVersion(ctx context.Context) (V, error) {
	if u.RemoteURL == "" {
		var zero V
		return zero, ErrRemoteURLNotSet
	}
	return u.VersionFunc(ctx, u.RemoteURL, u.RemoteFileRepository)
}

func (u *UpdatableFile[V]) Update(ctx context.Context, force bool) error {
	if need, err := u.needUpdate(ctx); !need || err != nil {
		return err
	}
	return u.update(ctx, force)
}

func (u *UpdatableFile[V]) update(ctx context.Context, force bool) error {
	if err := u.downloadTempFile(ctx, force); err != nil {
		return err
	}
	defer u.LocalFileRepository.Remove(ctx, u.tmpFilePath())

	if err := u.LocalFileRepository.Rename(ctx, u.tmpFilePath(), u.LocalPath); err != nil {
		return fmt.Errorf("failed to move temporary file: %w", err)
	}

	return nil
}

func (u *UpdatableFile[V]) updateInProgress(ctx context.Context) bool {
	exists, err := u.LocalFileRepository.Exists(ctx, u.tmpFilePath())
	if err != nil {
		return false
	}
	return exists
}

func (u *UpdatableFile[V]) downloadTempFile(ctx context.Context, force bool) error {
	if force {
		_ = u.LocalFileRepository.Remove(ctx, u.tmpFilePath())
	}

	tmpFile, err := u.LocalFileRepository.CreateIfNotExists(ctx, u.tmpFilePath())
	if err != nil {
		if errors.Is(err, ErrFileExists) {
			return utils.ErrUpdateInProgress
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
		if errors.Is(err, ErrRemoteURLNotSet) {
			return false, nil
		}
		return false, err
	}

	localVersion, err := u.Version(ctx)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return true, nil
		}
		return false, err
	}

	return remoteVersion.Compare(localVersion) > 0, nil
}

func (u *UpdatableFile[V]) tmpFilePath() string {
	return u.LocalPath + ".tmp"
}

func (u *UpdatableFile[V]) CheckUpdates(ctx context.Context) (upd Update[V], err error) {
	localVersion, err := u.Version(ctx)
	if err != nil {
		return upd, err
	}
	upd.CurrentVersion = localVersion

	remoteVersion, err := u.RemoteVersion(ctx)
	if err != nil {
		if errors.Is(err, ErrRemoteURLNotSet) {
			return upd, nil
		}
		return upd, err
	}

	upd.RemoteVersion = remoteVersion
	return upd, nil
}
