package maxmind

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"iter"
	"maps"
	"net/netip"
	"slices"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/bldsoft/geos/pkg/utils"
	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/utils/errgroup"
	"github.com/maxmind/mmdbwriter"
	"github.com/maxmind/mmdbwriter/inserter"
	"github.com/maxmind/mmdbwriter/mmdbtype"
	"github.com/oschwald/maxminddb-golang/v2"
)

var ErrNoDatabases = errors.New("no databases")

type mergedMMDB struct {
	key  string
	data []byte
}

type MultiMaxMindDB struct {
	dbs        []Database
	mu         sync.Mutex
	merged     *mergedMMDB
	mergeCount atomic.Uint32
}

func NewMultiMaxMindDB(dbs ...Database) *MultiMaxMindDB {
	res := &MultiMaxMindDB{dbs: dbs}
	return res
}

func (db *MultiMaxMindDB) Add(dbs ...Database) *MultiMaxMindDB {
	db.mu.Lock()
	db.dbs = append(db.dbs, dbs...)
	db.merged = nil
	db.mu.Unlock()
	return db
}

func (db *MultiMaxMindDB) invalidateMerged() {
	db.mu.Lock()
	db.merged = nil
	db.mu.Unlock()
}

func (db *MultiMaxMindDB) Lookup(ctx context.Context, ip netip.Addr, result any) error {
	ip = ip.Unmap()
	var multiErr error
	for _, v := range slices.Backward(db.dbs) {
		err := v.Lookup(ctx, ip, result)
		if err == nil {
			return nil
		}
		if !errors.Is(err, utils.ErrNotFound) {
			multiErr = errors.Join(multiErr, err)
		}
	}
	if multiErr != nil {
		return multiErr
	}
	return utils.ErrNotFound
}

func (db *MultiMaxMindDB) dbReader(ctx context.Context, index int) (*maxminddb.Reader, error) {
	database := db.dbs[index]
	reader, err := database.RawData(ctx)
	if err != nil {
		return nil, err
	}

	if buf, ok := reader.(*bytes.Buffer); ok {
		return maxminddb.OpenBytes(buf.Bytes())
	}

	bytes, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return maxminddb.OpenBytes(bytes)
}

func (db *MultiMaxMindDB) totalNodes(ctx context.Context) int {
	totalNodes := 0
	for _, db := range db.dbs {
		if meta, _ := db.MetaData(ctx); meta != nil {
			totalNodes += int(meta.NodeCount)
		}
	}
	return totalNodes
}

func (db *MultiMaxMindDB) nonEmptyDatabases(ctx context.Context) []Database {
	var res []Database
	for _, database := range db.dbs {
		if !db.isEmptyDatabase(ctx, database) {
			res = append(res, database)
		}
	}
	return res
}

func (db *MultiMaxMindDB) isEmptyDatabase(ctx context.Context, database Database) bool {
	meta, err := database.MetaData(ctx)
	if err != nil {
		return true
	}
	return meta.NodeCount == 0
}

func (db *MultiMaxMindDB) mergedCacheKey(ctx context.Context, dbs []Database) (string, bool) {
	var b strings.Builder
	for _, database := range dbs {
		meta, err := database.MetaData(ctx)
		if err != nil || meta == nil {
			return "", false
		}
		fmt.Fprintf(&b, "%s|%d|%d\n", meta.DatabaseType, meta.BuildEpoch, meta.NodeCount)
	}
	return b.String(), true
}

func (db *MultiMaxMindDB) cachedMerged(key string) []byte {
	db.mu.Lock()
	defer db.mu.Unlock()
	if db.merged == nil || db.merged.key != key {
		return nil
	}
	return append([]byte(nil), db.merged.data...)
}

func (db *MultiMaxMindDB) storeMerged(key string, data []byte) {
	db.mu.Lock()
	db.merged = &mergedMMDB{key: key, data: append([]byte(nil), data...)}
	db.mu.Unlock()
}

func (db *MultiMaxMindDB) RawData(ctx context.Context) (io.Reader, error) {
	nonEmtpyDbs := db.nonEmptyDatabases(ctx)
	if len(nonEmtpyDbs) == 0 {
		return nil, ErrNoDatabases
	}
	if len(nonEmtpyDbs) == 1 {
		return nonEmtpyDbs[0].RawData(ctx)
	}

	key, ok := db.mergedCacheKey(ctx, db.dbs)
	if ok {
		if cached := db.cachedMerged(key); cached != nil {
			return bytes.NewBuffer(cached), nil
		}
	}

	data, err := db.merge(ctx)
	if err != nil {
		return nil, err
	}
	if ok {
		db.storeMerged(key, data)
	}
	return bytes.NewBuffer(data), nil
}

func (db *MultiMaxMindDB) mergedBuildEpoch(base uint, patches ...uint) uint {
	epoch := base
	hasOlder := false
	for _, patch := range patches {
		if patch > epoch {
			epoch = patch
		} else if patch < base {
			hasOlder = true
		}
	}
	if hasOlder && epoch == base {
		epoch++
	}
	return epoch
}

func (db *MultiMaxMindDB) writerOptions(meta *maxminddb.Metadata) mmdbwriter.Options {
	desc := make(map[string]string, len(meta.Description))
	maps.Copy(desc, meta.Description)
	opts := mmdbwriter.Options{
		BuildEpoch:              int64(meta.BuildEpoch),
		DatabaseType:            meta.DatabaseType,
		Description:             desc,
		Languages:               append([]string(nil), meta.Languages...),
		IncludeReservedNetworks: true,
		DisableIPv4Aliasing:     true,
	}
	if meta.IPVersion != 0 {
		opts.IPVersion = int(meta.IPVersion)
	}
	if meta.RecordSize != 0 {
		opts.RecordSize = int(meta.RecordSize)
	}
	return opts
}

func (db *MultiMaxMindDB) merge(ctx context.Context) ([]byte, error) {
	db.mergeCount.Add(1)

	meta, err := db.MetaData(ctx)
	if err != nil {
		return nil, err
	}
	tree, err := mmdbwriter.New(db.writerOptions(meta))
	if err != nil {
		return nil, err
	}

	currentNode, totalNodes := 0, db.totalNodes(ctx)
	percent := (totalNodes / 100) + 1

	type networkNode struct {
		network netip.Prefix
		data    map[string]any
	}
	const bufSize = 1000
	readedNodeC := make(chan networkNode, bufSize)
	convertedNodeC := make(chan MMDBRecord, bufSize)

	var eg errgroup.Group
	eg.Go(func() (err error) {
		defer close(readedNodeC)
		for i := range db.dbs {
			dbReader, err := db.dbReader(ctx, i)
			if err != nil {
				return err
			}

			// SkipAliasedNetworks is now a default behavior
			for network := range dbReader.Networks() {
				if err := network.Err(); err != nil {
					return err
				}

				var node networkNode
				node.data = make(map[string]any)
				err := network.Decode(&node.data)
				if err != nil {
					return err
				}
				node.network = network.Prefix()
				readedNodeC <- node
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
			err = tree.InsertFunc(prefixToNetIPNet(convertedNode.Network), inserter.ReplaceWith(convertedNode.Data))
			if err != nil {
				log.FromContext(ctx).WarnWithFields(log.Fields{"err": err}, "failed to insert network")
				continue
			}
			currentNode++
			if currentNode%percent == 0 {
				percents := currentNode / percent
				log.FromContext(ctx).Debugf("Merging databases %d%%", percents)
			}
		}
		return nil
	})

	if err := eg.Wait(); err != nil {
		return nil, err
	}
	log.FromContext(ctx).Debug("Merging databases 100%")

	var buf bytes.Buffer
	if _, err := tree.WriteTo(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (db *MultiMaxMindDB) Reader(ctx context.Context) (*maxminddb.Reader, error) {
	reader, err := db.RawData(ctx)
	if err != nil {
		return nil, err
	}
	return maxminddb.OpenBytes(reader.(*bytes.Buffer).Bytes())
}

func (db *MultiMaxMindDB) Networks(ctx context.Context, options ...maxminddb.NetworksOption) (iter.Seq[maxminddb.Result], error) {
	reader, err := db.Reader(ctx)
	if err != nil {
		return nil, err
	}
	return reader.Networks(options...), nil
}

func (db *MultiMaxMindDB) MetaData(ctx context.Context) (*maxminddb.Metadata, error) {
	if len(db.dbs) == 0 {
		return nil, ErrNoDatabases
	}
	if len(db.dbs) == 1 {
		return db.dbs[0].MetaData(ctx)
	}

	mainMeta, err := db.dbs[0].MetaData(ctx)
	if err != nil {
		return nil, err
	}

	res := *mainMeta
	res.Description = make(map[string]string)
	maps.Copy(res.Description, mainMeta.Description)
	res.Description["en"] += " patched by GEOS service."

	var patchEpochs []uint
	for _, child := range db.dbs[1:] {
		meta, err := child.MetaData(ctx)
		if err != nil {
			log.FromContext(ctx).WarnWithFields(log.Fields{"err": err}, "failed to get db metadata")
			continue
		}
		patchEpochs = append(patchEpochs, meta.BuildEpoch)
	}
	res.BuildEpoch = db.mergedBuildEpoch(mainMeta.BuildEpoch, patchEpochs...)

	return &res, nil
}
