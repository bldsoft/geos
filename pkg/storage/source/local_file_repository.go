package source

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"time"
)

type LocalFileRepository struct {
}

func NewLocalFileRepository() *LocalFileRepository {
	return &LocalFileRepository{}
}

func (r *LocalFileRepository) Reader(ctx context.Context, path string) (io.ReadCloser, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrFileNotExists
		}
		return nil, err
	}
	return file, nil
}

func (r *LocalFileRepository) TailReader(ctx context.Context, path string, offset int64) (io.ReadCloser, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	fileInfo, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, err
	}

	fileSize := fileInfo.Size()
	startOffset := fileSize - offset
	if startOffset < 0 {
		startOffset = 0
	}

	_, err = file.Seek(startOffset, io.SeekStart)
	if err != nil {
		file.Close()
		return nil, err
	}

	return file, nil
}

func (r *LocalFileRepository) LastModified(ctx context.Context, path string) (time.Time, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return time.Time{}, nil
		}
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

func (r *LocalFileRepository) TryLock(ctx context.Context, path string) (ok bool, close func(), err error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return false, nil, err
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0644)
	if err != nil {
		if os.IsExist(err) {
			return false, nil, nil
		}
		return false, nil, err
	}
	return true, func() {
		file.Close()
		os.Remove(path)
	}, nil
}

func (r *LocalFileRepository) Open(ctx context.Context, path string) (io.WriteCloser, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}
	return os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0644)
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
