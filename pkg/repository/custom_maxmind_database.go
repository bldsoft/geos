package repository

import (
	"bytes"
	"errors"
	"io"
	"net"

	"github.com/maxmind/mmdbwriter"
	"github.com/maxmind/mmdbwriter/inserter"
	"github.com/maxmind/mmdbwriter/mmdbtype"
	"github.com/oschwald/maxminddb-golang"
)

type MMDBRecord struct {
	Network *net.IPNet
	Data    mmdbtype.Map
}

type MMDBRecordReader interface {
	ReadMMDBRecord() (MMDBRecord, error)
}

type CustomMaxMindDB struct {
	tree  *mmdbwriter.Tree
	dbRaw []byte
	db    *maxminddb.Reader
}

func NewCustomMaxMindDB(reader MMDBRecordReader) (*CustomMaxMindDB, error) {
	tree, err := mmdbwriter.New(mmdbwriter.Options{IncludeReservedNetworks: true})
	if err != nil {
		return nil, err
	}

	for {
		rec, err := reader.ReadMMDBRecord()
		if err != nil {
			break
		}
		if err := tree.InsertFunc(rec.Network, inserter.TopLevelMergeWith(rec.Data)); err != nil {
			return nil, err
		}
	}

	if err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}
	err = nil

	var buf bytes.Buffer
	if _, err := tree.WriteTo(&buf); err != nil {
		return nil, err
	}
	dbRaw := buf.Bytes()

	db, err := maxminddb.FromBytes(buf.Bytes())
	if err != nil {
		return nil, err
	}
	return &CustomMaxMindDB{
		tree:  tree,
		dbRaw: dbRaw,
		db:    db,
	}, nil
}

func (db *CustomMaxMindDB) Available() bool {
	return true
}

func (db *CustomMaxMindDB) Lookup(ip net.IP, result interface{}) error {
	return db.db.Lookup(ip, result)
}

func (db *CustomMaxMindDB) Networks(options ...maxminddb.NetworksOption) (*maxminddb.Networks, error) {
	return db.db.Networks(options...), nil
}

func (db *CustomMaxMindDB) RawData() (io.Reader, error) {
	return bytes.NewReader(db.dbRaw), nil
}

func (db *CustomMaxMindDB) MetaData() (*maxminddb.Metadata, error) {
	return &db.db.Metadata, nil
}
