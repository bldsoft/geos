package utils

import (
	"errors"
	"io"
)

type teeReadCloser struct {
	r io.ReadCloser
	w io.WriteCloser
}

func NewTeeReadCloser(r io.ReadCloser, w io.WriteCloser) *teeReadCloser {
	return &teeReadCloser{r, w}
}

func (t *teeReadCloser) Read(p []byte) (n int, err error) {
	n, err = t.r.Read(p)
	if n > 0 {
		if n, err := t.w.Write(p[:n]); err != nil {
			return n, err
		}
	}
	return
}

func (t *teeReadCloser) Close() error {
	err := t.r.Close()
	err = errors.Join(err, t.w.Close())
	return err
}
