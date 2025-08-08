package utils

import (
	"archive/tar"
	"io"

	"github.com/klauspost/compress/gzip"
)

// empty archive is not an error, but an empty map
func UnpackTarGz(src io.Reader) (map[string][]byte, error) {
	gr, err := gzip.NewReader(src)
	if err != nil {
		if err == io.EOF {
			return make(map[string][]byte), nil
		}

		return nil, err
	}
	defer gr.Close()

	tr := tar.NewReader(gr)
	files := make(map[string][]byte)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if header.Typeflag == tar.TypeReg {
			data, err := io.ReadAll(tr)
			if err != nil {
				return nil, err
			}
			files[header.Name] = data
		}
	}
	return files, nil
}
