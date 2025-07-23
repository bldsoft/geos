package maxmind

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"path/filepath"

	"github.com/bldsoft/geos/pkg/entity"
	"github.com/bldsoft/geos/pkg/storage/source"
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

func NewDatabasePatchesFromTarGz(source *source.PatchesSource) ([]*DatabasePatch, error) {
	ctx := context.Background()
	r, err := source.Reader(ctx)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	contents, err := utils.UnpackTarGz(r)
	if err != nil {
		return nil, err
	}

	var customDBs []*DatabasePatch
	for fileName, content := range contents {
		if filepath.Ext(fileName) != ".json" {
			return nil, utils.ErrUnknownFormat
		}

		jsonReader, err := NewJSONRecordReader(bytes.NewReader(content))
		if err != nil {
			return nil, err
		}

		db, err := NewDatabasePatch(jsonReader)
		if err != nil {
			return nil, err
		}

		ver, err := source.Version(ctx)
		if err != nil {
			return nil, err
		}

		customDBs = append(customDBs, db.WithMetadata(maxminddb.Metadata{
			Description:              map[string]string{"en": fmt.Sprintf("path = %s", fileName)},
			DatabaseType:             db.db.Metadata.DatabaseType,
			Languages:                db.db.Metadata.Languages,
			BinaryFormatMajorVersion: db.db.Metadata.BinaryFormatMajorVersion,
			BinaryFormatMinorVersion: db.db.Metadata.BinaryFormatMinorVersion,
			BuildEpoch:               uint(ver.Time().Unix()),
			IPVersion:                db.db.Metadata.IPVersion,
			NodeCount:                db.db.Metadata.NodeCount,
			RecordSize:               db.db.Metadata.RecordSize,
		}).WithState(ver.Time().Unix()))
	}

	return customDBs, nil
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
	return nil, errors.ErrUnsupported
}

func (db *DatabasePatch) CheckUpdates(_ context.Context) (entity.Updates, error) {
	return nil, errors.ErrUnsupported
}

func (db *DatabasePatch) LastUpdateInterrupted(_ context.Context) (bool, error) {
	return false, errors.ErrUnsupported
}

func (db *DatabasePatch) State() *state.GeosState {
	return &state.GeosState{}
}
