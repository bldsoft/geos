package maxmind

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"os"

	"github.com/bldsoft/geos/pkg/entity"
	"github.com/bldsoft/geos/pkg/storage/source"
	"github.com/oschwald/maxminddb-golang"
)

var ErrNoSource = fmt.Errorf("no source provided for the database")

type MaxmindDatabase struct {
	path   string
	reader *maxminddb.Reader
	dbRaw  []byte
	source *source.MaxmindSource
}

func Open(path string) (*MaxmindDatabase, error) {
	dbRaw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	db, err := maxminddb.FromBytes(dbRaw)
	if err != nil {
		return nil, err
	}

	return &MaxmindDatabase{
		path:   path,
		reader: db,
		dbRaw:  dbRaw,
	}, nil
}

func (db *MaxmindDatabase) SetSource(source *source.MaxmindSource) {
	db.source = source
}

func (db *MaxmindDatabase) Lookup(ip net.IP, result interface{}) error {
	return db.reader.Lookup(ip, result)
}

func (db *MaxmindDatabase) RawData() (io.Reader, error) {
	return bytes.NewBuffer(db.dbRaw), nil
}

func (db *MaxmindDatabase) MetaData() (*maxminddb.Metadata, error) {
	return &db.reader.Metadata, nil
}

func (db *MaxmindDatabase) Networks(options ...maxminddb.NetworksOption) (*maxminddb.Networks, error) {
	return db.reader.Networks(), nil
}

func (db *MaxmindDatabase) Download(ctx context.Context, update ...bool) (entity.Updates, error) {
	if db.source == nil {
		return nil, ErrNoSource
	}

	return db.source.Download(ctx, update...)
}

func (db *MaxmindDatabase) CheckUpdates(ctx context.Context) (entity.Updates, error) {
	if db.source == nil {
		return nil, ErrNoSource
	}
	return db.source.CheckUpdates(ctx)
}
