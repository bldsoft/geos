package repository

import (
	"bytes"
	"io"
	"os"

	"github.com/oschwald/maxminddb-golang"
)

type database struct {
	path string
	*maxminddb.Reader
	dbRaw []byte
}

func openDB(path string) (*database, error) {
	dbRaw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	db, err := maxminddb.FromBytes(dbRaw)
	if err != nil {
		return nil, err
	}

	return &database{
		path:   path,
		Reader: db,
		dbRaw:  dbRaw,
	}, nil
}

func (db *database) RawData() (io.Reader, error) {
	return bytes.NewBuffer(db.dbRaw), nil
}

func (db *database) MetaData() (*maxminddb.Metadata, error) {
	return &db.Reader.Metadata, nil
}

func (db *database) Networks(options ...maxminddb.NetworksOption) (*maxminddb.Networks, error) {
	return db.Reader.Networks(), nil
}
