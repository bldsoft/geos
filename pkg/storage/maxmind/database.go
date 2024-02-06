package maxmind

import (
	"bytes"
	"io"
	"net"
	"os"

	"github.com/oschwald/maxminddb-golang"
)

type MaxmindDatabase struct {
	path   string
	reader *maxminddb.Reader
	dbRaw  []byte
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
