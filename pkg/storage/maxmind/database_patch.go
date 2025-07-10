package maxmind

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"

	"github.com/bldsoft/geos/pkg/entity"
	"github.com/bldsoft/geos/pkg/storage/state"
	"github.com/bldsoft/geos/pkg/utils"
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

type DatabasePatch struct {
	tree  *mmdbwriter.Tree
	dbRaw []byte
	db    *maxminddb.Reader
	state int64
}

func NewDatabasePatchesFromTarGz(filename string) ([]*DatabasePatch, error) {
	var customDBs []*DatabasePatch
	file, err := os.OpenFile(filename, os.O_RDONLY, 0666)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	contents, err := utils.UnpackTarGz(file)
	if err != nil {
		return nil, err
	}

	for fileName, content := range contents {
		if filepath.Ext(fileName) != ".json" {
			return nil, utils.ErrUnknownFormat
		}

		reader, err := NewJSONRecordReader(bytes.NewReader(content))
		if err != nil {
			return nil, err
		}

		if db, err := NewDatabasePatch(reader); err != nil {
			return nil, err
		} else {
			customDBs = append(customDBs, db.WithMetadata(maxminddb.Metadata{
				Description:              map[string]string{"en": fmt.Sprintf("path = %s", fileName)},
				DatabaseType:             db.db.Metadata.DatabaseType,
				Languages:                db.db.Metadata.Languages,
				BinaryFormatMajorVersion: db.db.Metadata.BinaryFormatMajorVersion,
				BinaryFormatMinorVersion: db.db.Metadata.BinaryFormatMinorVersion,
				BuildEpoch:               uint(stat.ModTime().Unix()),
				IPVersion:                db.db.Metadata.IPVersion,
				NodeCount:                db.db.Metadata.NodeCount,
				RecordSize:               db.db.Metadata.RecordSize,
			}).WithState(stat.ModTime().Unix()))
		}
	}

	return customDBs, nil
}

func NewDatabasePatchFromFile(path string) (*DatabasePatch, error) {
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
		return nil, utils.ErrUnknownFormat
	}
	if err != nil {
		return nil, err
	}

	db, err := NewDatabasePatch(reader)
	if err != nil {
		return nil, err
	}

	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}
	return db.WithMetadata(maxminddb.Metadata{
		Description:              map[string]string{"en": fmt.Sprintf("path = %s", path)},
		DatabaseType:             db.db.Metadata.DatabaseType,
		Languages:                db.db.Metadata.Languages,
		BinaryFormatMajorVersion: db.db.Metadata.BinaryFormatMajorVersion,
		BinaryFormatMinorVersion: db.db.Metadata.BinaryFormatMinorVersion,
		BuildEpoch:               uint(stat.ModTime().Unix()),
		IPVersion:                db.db.Metadata.IPVersion,
		NodeCount:                db.db.Metadata.NodeCount,
		RecordSize:               db.db.Metadata.RecordSize,
	}).WithState(stat.ModTime().Unix()), nil
}

func NewDatabasePatch(reader MMDBRecordReader) (*DatabasePatch, error) {
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

	var buf bytes.Buffer
	if _, err := tree.WriteTo(&buf); err != nil {
		return nil, err
	}
	dbRaw := buf.Bytes()

	db, err := maxminddb.FromBytes(buf.Bytes())
	if err != nil {
		return nil, err
	}
	return &DatabasePatch{
		tree:  tree,
		dbRaw: dbRaw,
		db:    db,
	}, nil
}

func (db *DatabasePatch) WithMetadata(meta maxminddb.Metadata) *DatabasePatch {
	db.db.Metadata = meta
	return db
}

func (db *DatabasePatch) WithState(state int64) *DatabasePatch {
	db.state = state
	return db
}

func (db *DatabasePatch) Available() bool {
	return true
}

func (db *DatabasePatch) Lookup(ip net.IP, result interface{}) error {
	_, ok, err := db.db.LookupNetwork(ip, result)
	if err != nil {
		return err
	}
	if !ok {
		return utils.ErrNotFound
	}
	return nil
}

func (db *DatabasePatch) Networks(options ...maxminddb.NetworksOption) (*maxminddb.Networks, error) {
	return db.db.Networks(options...), nil
}

func (db *DatabasePatch) RawData() (io.Reader, error) {
	return bytes.NewReader(db.dbRaw), nil
}

func (db *DatabasePatch) MetaData() (*maxminddb.Metadata, error) {
	return &db.db.Metadata, nil
}

//--- these are controlled by the custom database

func (db *DatabasePatch) Download(_ context.Context) (entity.Updates, error) {
	return nil, nil
}

func (db *DatabasePatch) CheckUpdates(_ context.Context) (entity.Updates, error) {
	return nil, nil
}

func (db *DatabasePatch) State() *state.GeosState {
	return &state.GeosState{}
}
