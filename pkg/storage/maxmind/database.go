package maxmind

import (
	"bytes"
	"io"
	"os"

	"github.com/oschwald/maxminddb-golang"
)

type MaxmindDatabase struct {
	path string
	*maxminddb.Reader
	dbRaw []byte
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
		Reader: db,
		dbRaw:  dbRaw,
	}, nil
}

func (db *MaxmindDatabase) RawData() (io.Reader, error) {
	return bytes.NewBuffer(db.dbRaw), nil
}

func (db *MaxmindDatabase) MetaData() (*maxminddb.Metadata, error) {
	return &db.Reader.Metadata, nil
}

func (db *MaxmindDatabase) Networks(options ...maxminddb.NetworksOption) (*maxminddb.Networks, error) {
	return db.Reader.Networks(), nil
}