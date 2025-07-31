package maxmind

import (
	"bytes"
	"errors"
	"io"
	"net"

	"maps"

	"github.com/bldsoft/geos/pkg/utils"
	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/utils/errgroup"
	"github.com/maxmind/mmdbwriter"
	"github.com/maxmind/mmdbwriter/inserter"
	"github.com/maxmind/mmdbwriter/mmdbtype"
	"github.com/oschwald/maxminddb-golang"
)

var ErrNoDatabases = errors.New("no databases")

type MultiMaxMindDB struct {
	dbs    []Database
	logger log.ServiceLogger
}

func NewMultiMaxMindDB(dbs ...Database) *MultiMaxMindDB {
	res := &MultiMaxMindDB{dbs: dbs, logger: log.Logger}
	return res
}

func (db *MultiMaxMindDB) WithLogger(logger log.ServiceLogger) *MultiMaxMindDB {
	db.logger = logger
	return db
}

func (db *MultiMaxMindDB) Add(dbs ...Database) *MultiMaxMindDB {
	db.dbs = append(db.dbs, dbs...)
	return db
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

	if buf, ok := reader.(*bytes.Buffer); ok {
		return maxminddb.FromBytes(buf.Bytes())
	}

	bytes, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return maxminddb.FromBytes(bytes)
}

func (db *MultiMaxMindDB) totalNodes() int {
	totalNodes := 0
	for _, db := range db.dbs {
		if meta, _ := db.MetaData(); meta != nil {
			totalNodes += int(meta.NodeCount)
		}
	}
	return totalNodes
}

func (db *MultiMaxMindDB) nonEmptyDatabases() []Database {
	var res []Database
	for _, database := range db.dbs {
		if !db.isEmptyDatabase(database) {
			res = append(res, database)
		}
	}
	return res
}

func (db *MultiMaxMindDB) isEmptyDatabase(database Database) bool {
	meta, err := database.MetaData()
	if err != nil {
		return true
	}
	return meta.NodeCount == 0
}

func (db *MultiMaxMindDB) RawData() (io.Reader, error) {
	nonEmtpyDbs := db.nonEmptyDatabases()
	if len(nonEmtpyDbs) == 0 {
		return nil, ErrNoDatabases
	}
	if len(nonEmtpyDbs) == 1 {
		return nonEmtpyDbs[0].RawData()
	}

	opts := mmdbwriter.Options{IncludeReservedNetworks: true}
	tree, err := mmdbwriter.New(opts)
	if err != nil {
		return nil, err
	}

	currentNode, totalNodes := 0, db.totalNodes()
	percent := (totalNodes / 100) + 1

	type networkNode struct {
		network *net.IPNet
		data    map[string]interface{}
	}
	const bufSize = 1000
	readedNodeC := make(chan networkNode, bufSize)
	convertedNodeC := make(chan MMDBRecord, bufSize)

	var eg errgroup.Group
	eg.Go(func() (err error) {
		defer close(readedNodeC)
		for i := range db.dbs {
			dbReader, err := db.dbReader(i)
			if err != nil {
				return err
			}
			networks := dbReader.Networks(maxminddb.SkipAliasedNetworks)
			for networks.Next() {
				var node networkNode
				node.data = make(map[string]interface{})
				node.network, err = networks.Network(&node.data)
				if err != nil {
					return err
				}
				readedNodeC <- node
			}
			if err := networks.Err(); err != nil {
				return err
			}
		}
		return nil
	})

	eg.Go(func() (err error) {
		defer close(convertedNodeC)
		for node := range readedNodeC {
			var rec MMDBRecord
			rec.Network = node.network
			rec.Data, _ = toMMDBType(node.data).(mmdbtype.Map)
			convertedNodeC <- rec
		}
		return nil
	})

	eg.Go(func() error {
		for convertedNode := range convertedNodeC {
			err = tree.InsertFunc(convertedNode.Network, inserter.ReplaceWith(convertedNode.Data))
			if err != nil {
				return err
			}
			currentNode++
			if currentNode%percent == 0 {
				percents := currentNode / percent
				db.logger.Debugf("Merging databases %d%%", percents)
			}
		}
		return nil
	})

	if err := eg.Wait(); err != nil {
		return nil, err
	}
	db.logger.Debug("Merging databases 100%")

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
		return nil, ErrNoDatabases
	}
	if len(db.dbs) == 1 {
		return db.dbs[0].MetaData()
	}

	mainMeta, err := db.dbs[0].MetaData()
	if err != nil {
		return nil, err
	}

	res := *mainMeta
	res.Description = make(map[string]string)
	maps.Copy(res.Description, mainMeta.Description)
	res.Description["en"] += " patched by GEOS service."

	for _, db := range db.dbs {
		meta, err := db.MetaData()
		if err != nil {
			log.Logger.WarnWithFields(log.Fields{"err": err}, "failed to get db metadata")
			continue
		}
		res.BuildEpoch = max(res.BuildEpoch, meta.BuildEpoch)
	}

	return &res, nil
}
