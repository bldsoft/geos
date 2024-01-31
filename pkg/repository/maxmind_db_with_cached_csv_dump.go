package repository

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bldsoft/geos/pkg/storage/maxmind"
	"github.com/bldsoft/gost/log"
	"github.com/oschwald/maxminddb-golang"
)

type maxmindDBWithCachedCSVDump struct {
	maxmind.CSVDumper
	archivedCSVWithNamesDump []byte
	dumpReady                chan struct{}
}

func withCachedCSVDump[T maxmind.CSVEntity](db maxmind.Database) *maxmindDBWithCachedCSVDump {
	return &maxmindDBWithCachedCSVDump{CSVDumper: maxmind.NewCSVDumper[T](db), dumpReady: make(chan struct{})}
}

func (db *maxmindDBWithCachedCSVDump) initCSVDump(ctx context.Context, csvDumpPath string) {
	defer close(db.dumpReady)
	if filepath.Dir(csvDumpPath) == "." {
		return
	}

	// TODO: remove
	_ = os.Remove(csvDumpPath)          // remove old uncompressed file
	_ = os.Remove(csvDumpPath + ".tmp") // remove old uncompressed file
	// =============

	csvDumpPath = csvDumpPath + ".gz"

	dir := filepath.Dir(csvDumpPath)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		panic(fmt.Errorf("failed to create dir for csv dump: %w", err))
	}

	var err error
	db.archivedCSVWithNamesDump, err = db.csvDumpFromDisk(ctx, csvDumpPath)
	if err != nil {
		panic(fmt.Errorf("failed to load GeoIP dump: %w", err))
	}
	log.FromContext(ctx).InfoWithFields(log.Fields{"path": csvDumpPath, "size MB": len(db.archivedCSVWithNamesDump) / 1024 / 1024}, "Dump loaded to memory")
}

func (db *maxmindDBWithCachedCSVDump) metadata() *maxminddb.Metadata {
	meta, _ := db.MetaData()
	return meta
}

func (db *maxmindDBWithCachedCSVDump) csvDumpFromDisk(ctx context.Context, dumpPath string) ([]byte, error) {
	needUpdate, err := db.needUpdateDump(ctx, dumpPath)
	if err != nil {
		return nil, err
	}
	if !needUpdate {
		return os.ReadFile(dumpPath)
	}

	log.FromContext(ctx).InfoWithFields(log.Fields{"meta": db.metadata(), "csv": dumpPath}, "Updating CSV")
	return db.loadDumpFull(ctx, dumpPath)
}

func (db *maxmindDBWithCachedCSVDump) needUpdateDump(ctx context.Context, dumpPath string) (bool, error) {
	_, err := os.Stat(dumpPath)
	if err != nil {
		if os.IsNotExist(err) {
			return true, nil
		}
		return false, err
	}

	dumpMetaData, err := db.dumpMetaData(dumpPath)
	if err != nil {
		log.FromContext(ctx).DebugWithFields(log.Fields{"err": err}, "Failed to get dump metadata")
		return true, nil
	}
	dumpDBBuildTime := time.Unix(int64(dumpMetaData.BuildEpoch), 0)
	dbBuildTime := time.Unix(int64(db.metadata().BuildEpoch), 0)
	return dbBuildTime.After(dumpDBBuildTime), nil
}

func (db *maxmindDBWithCachedCSVDump) loadDumpFull(ctx context.Context, dumpPath string) ([]byte, error) {
	temp := dumpPath + ".tmp"

	var buf bytes.Buffer
	file, err := os.Create(temp)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	w := io.MultiWriter(&buf, file)

	err = func() error {
		gw := gzip.NewWriter(w)
		defer gw.Close()
		return db.CSVDumper.WriteCSVTo(ctx, gw)
	}()
	if err != nil {
		return nil, os.Remove(temp)
	}

	if err := db.writeDumpMetadata(dumpPath); err != nil {
		return nil, err
	}

	return buf.Bytes(), os.Rename(temp, dumpPath)
}

func (db *maxmindDBWithCachedCSVDump) dumpMetaDataPath(dumpPath string) string {
	ext := filepath.Ext(dumpPath)
	return strings.TrimSuffix(dumpPath, ext) + ".meta"
}

func (db *maxmindDBWithCachedCSVDump) dumpMetaData(dumpPath string) (*maxminddb.Metadata, error) {
	dumpMetaDataPath := db.dumpMetaDataPath(dumpPath)

	data, err := os.ReadFile(dumpMetaDataPath)
	if err != nil {
		return nil, err
	}

	var meta maxminddb.Metadata
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, err
	}

	return &meta, nil
}

func (db *maxmindDBWithCachedCSVDump) writeDumpMetadata(dumpPath string) error {
	dumpMetaDataPath := db.dumpMetaDataPath(dumpPath)
	data, err := json.Marshal(db.metadata())
	if err != nil {
		return err
	}
	return os.WriteFile(dumpMetaDataPath, data, 0666)
}

func (db *maxmindDBWithCachedCSVDump) CSV(ctx context.Context, gzipCompress bool) (io.Reader, error) {
	select {
	case <-db.dumpReady:
		res := bytes.NewReader(db.archivedCSVWithNamesDump)
		if gzipCompress {
			return res, nil
		}
		return gzip.NewReader(res)
	default:
		return nil, ErrGeoIPCSVNotReady
	}
}
