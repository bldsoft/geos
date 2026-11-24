package source

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

var sameHostHTTPClient = &http.Client{
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= 10 {
			return errors.New("stopped after 10 redirects")
		}
		if req.URL.Hostname() != via[0].URL.Hostname() {
			return fmt.Errorf("refusing redirect to %s", req.URL.Host)
		}
		return nil
	},
}

type RemoteFileRepository struct {
	client *http.Client
}

func NewRemoteFileRepository() *RemoteFileRepository {
	return &RemoteFileRepository{client: http.DefaultClient}
}

func (r *RemoteFileRepository) Reader(ctx context.Context, path string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("server did not return OK: %s", resp.Status)
	}

	return resp.Body, nil
}

func (r *RemoteFileRepository) TailReader(ctx context.Context, path string, offset int64) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Range", fmt.Sprintf("bytes=-%d", offset))

	resp, err := r.client.Do(req)
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

	resp, err := r.client.Do(req)
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
