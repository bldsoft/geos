package source

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

type RemoteFileRepository struct {
}

func NewRemoteFileRepository() *RemoteFileRepository {
	return &RemoteFileRepository{}
}

func (r *RemoteFileRepository) Reader(ctx context.Context, path string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

func (r *RemoteFileRepository) TailReader(ctx context.Context, path string, offset int64) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Range", fmt.Sprintf("bytes=-%d", offset))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusPartialContent {
		resp.Body.Close()
		return nil, fmt.Errorf("server did not return partial content: %s", resp.Status)
	}

	return resp.Body, nil
}

func (r *RemoteFileRepository) LastModified(ctx context.Context, path string) (time.Time, error) {
	req, err := http.NewRequestWithContext(ctx, "HEAD", path, nil)
	if err != nil {
		return time.Time{}, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return time.Time{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return time.Time{}, fmt.Errorf("server did not return OK: %s", resp.Status)
	}

	lastModified, err := time.Parse(http.TimeFormat, resp.Header.Get("Last-Modified"))
	if err != nil {
		return time.Time{}, err
	}

	return lastModified, nil
}

func (r *RemoteFileRepository) Exists(ctx context.Context, path string) (bool, error) {
	return true, nil
}

var _ ReadFileRepository = &RemoteFileRepository{}
