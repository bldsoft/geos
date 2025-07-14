package source2

import (
	"context"
	"io"
	"os"
	"time"
)

type LocalFileRepository struct {
}

func NewLocalFileRepository() *LocalFileRepository {
	return &LocalFileRepository{}
}

func (r *LocalFileRepository) Reader(ctx context.Context, path string) (io.ReadCloser, error) {
	return os.Open(path)
}

func (r *LocalFileRepository) TailReader(ctx context.Context, path string, offset int64) (io.ReadCloser, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	_, err = file.Seek(offset, io.SeekEnd)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func (r *LocalFileRepository) LastModified(ctx context.Context, path string) (time.Time, error) {
	info, err := os.Stat(path)
	if err != nil {
		return time.Time{}, err
	}
	return info.ModTime(), nil
}

func (r *LocalFileRepository) Exists(ctx context.Context, path string) (bool, error) {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *LocalFileRepository) CreateIfNotExists(ctx context.Context, path string) (io.WriteCloser, error) {
	return os.OpenFile(path, os.O_CREATE|os.O_EXCL, 0644)
}

func (r *LocalFileRepository) Write(ctx context.Context, path string, reader io.Reader) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, reader)
	return err
}

func (r *LocalFileRepository) Rename(ctx context.Context, oldPath, newPath string) error {
	return os.Rename(oldPath, newPath)
}

func (r *LocalFileRepository) Remove(ctx context.Context, path string) error {
	return os.Remove(path)
}

var _ FileRepository = &LocalFileRepository{}
