package repository

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/bldsoft/gost/log"
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

func NewCustomDatabasesFromDir(dir, customDBPrefix string) []maxmindDatabase {
	var customDBs []maxmindDatabase
	err := filepath.WalkDir(dir, func(s string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if !d.Type().IsRegular() && !strings.HasPrefix(d.Name(), customDBPrefix) {
			return nil
		}

		path := filepath.Join(dir, d.Name())
		custom, err := NewCustomMaxMindDBFromFile(path)
		if err != nil {
			if errors.Is(err, ErrUnknownFormat) {
				return nil
			}
			return err
		}
		customDBs = append(customDBs, custom)

		return nil
	})
	if err != nil {
		log.WarnWithFields(log.Fields{"err": err}, "failed to read custom databases")
	}
	return customDBs
}

func NewCustomMaxMindDBFromFile(path string) (*CustomMaxMindDB, error) {
	file, err := os.OpenFile(path, os.O_RDONLY, 0666)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var reader MMDBRecordReader
	switch filepath.Ext(path) {
	case ".json":
		reader, err = NewJSONRecordReader(file)
	default:
		return nil, ErrUnknownFormat
	}
	if err != nil {
		return nil, err
	}

	return NewCustomMaxMindDB(reader)
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
