package repository

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"io"
	"net"

	"github.com/bldsoft/geos/pkg/utils"
	"github.com/maxmind/mmdbwriter"
	"github.com/maxmind/mmdbwriter/inserter"
	"github.com/maxmind/mmdbwriter/mmdbtype"
	"github.com/oschwald/maxminddb-golang"
)

type MultiMaxMindDB struct {
	dbs []maxmindDatabase
}

func NewMultiMaxMindDB(dbs ...maxmindDatabase) *MultiMaxMindDB {
	res := &MultiMaxMindDB{dbs: dbs}
	return res
}

func (db *MultiMaxMindDB) Add(dbs ...maxmindDatabase) *MultiMaxMindDB {
	db.dbs = append(db.dbs, dbs...)
	return db
}

func (db *MultiMaxMindDB) Available() bool {
	for _, db := range db.dbs {
		if !db.Available() {
			return false
		}
	}
	return true
}

func (db *MultiMaxMindDB) Lookup(ip net.IP, result interface{}) error {
	var multiErr error
	for i := len(db.dbs) - 1; i >= 0; i-- {
		err := db.dbs[i].Lookup(ip, result)
		if err == nil {
			return nil
		}
		multiErr = errors.Join(multiErr, err)
	}
	return errors.Join(utils.ErrNotFound, multiErr)
}

func (db *MultiMaxMindDB) dbReader(index int) (*maxminddb.Reader, error) {
	database := db.dbs[index]
	reader, err := database.RawData()
	if err != nil {
		return nil, err
	}

	bytes, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	return maxminddb.FromBytes(bytes)
}

func (db *MultiMaxMindDB) RawData() (io.Reader, error) {
	opts := mmdbwriter.Options{IncludeReservedNetworks: true}
	tree, err := mmdbwriter.New(opts)
	if err != nil {
		return nil, err
	}

	for i := range db.dbs {
		dbReader, err := db.dbReader(i)
		if err != nil {
			return nil, err
		}

		networks := dbReader.Networks(maxminddb.SkipAliasedNetworks)

		for networks.Next() {
			var rec MMDBRecord

			m := make(map[string]interface{})
			network, err := networks.Network(&m)
			if err != nil {
				return nil, err
			}
			rec.Data = toMMDBType(m).(mmdbtype.Map)

			err = tree.InsertFunc(network, inserter.ReplaceWith(rec.Data))
			if err != nil {
				return nil, err
			}
		}

		if err := networks.Err(); err != nil {
			return nil, err
		}
	}

	var buf bytes.Buffer
	if _, err := tree.WriteTo(&buf); err != nil {
		return nil, err
	}
	return &buf, nil
}

func (db *MultiMaxMindDB) Reader() (*maxminddb.Reader, error) {
	reader, err := db.RawData()
	if err != nil {
		return nil, err
	}
	return maxminddb.FromBytes(reader.(*bytes.Buffer).Bytes())
}

func (db *MultiMaxMindDB) Networks(options ...maxminddb.NetworksOption) (*maxminddb.Networks, error) {
	reader, err := db.Reader()
	if err != nil {
		return nil, err
	}
	return reader.Networks(options...), nil
}

func (db *MultiMaxMindDB) MetaData() (*maxminddb.Metadata, error) {
	if len(db.dbs) == 0 {
		return nil, errors.New("no databases")
	}
	if len(db.dbs) == 1 {
		return db.dbs[0].MetaData()
	}

	mainMeta, err := db.dbs[0].MetaData()
	if err != nil {
		return nil, err
	}

	var res maxminddb.Metadata
	res = *mainMeta
	res.Description = make(map[string]string)
	for key, value := range mainMeta.Description {
		res.Description[key] = value
	}
	res.Description["en"] += "Database patched by geos service."

	return &res, nil
}

func (db *MultiMaxMindDB) WriteCSVTo(ctx context.Context, w io.Writer) error {
	for _, db := range db.dbs {
		if dumper, ok := db.(csvDumper); ok {
			if err := dumper.WriteCSVTo(ctx, w); err != nil {
				return err
			}
		}
	}
	return nil
}

func (db *MultiMaxMindDB) CSV(ctx context.Context, gzipCompress bool) (io.Reader, error) {
	var buf bytes.Buffer
	var w io.Writer = &buf
	if gzipCompress {
		w = gzip.NewWriter(w)
	}
	db.WriteCSVTo(ctx, w)
	return &buf, nil
}
