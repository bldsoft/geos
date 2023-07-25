package repository

import (
	"bytes"
	"context"
	"io"

	"github.com/bldsoft/gost/utils"
)

type ispDB struct {
	*database
}

func openISPDB(path string, required bool) *ispDB {
	return &ispDB{openDB(path, MaxmindDBTypeISP, required)}
}

func (db *ispDB) WriteCSVTo(ctx context.Context, w io.Writer) error {
	return utils.ErrNotImplemented
}

func (db *ispDB) CSV(ctx context.Context, withColumnNames bool) ([]byte, error) {
	var buf bytes.Buffer
	db.WriteCSVTo(ctx, &buf)
	return buf.Bytes(), nil
}
