package repository

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/bldsoft/gost/log"
)

type maxmindDBWithCachedCSVDump struct {
	maxmindCSVDumper
	archivedCSVWithNamesDump []byte
	dumpReady                chan struct{}
}

func withCachedCSVDump(db maxmindCSVDumper) *maxmindDBWithCachedCSVDump {
	return &maxmindDBWithCachedCSVDump{maxmindCSVDumper: db, dumpReady: make(chan struct{})}
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

	if !db.Available() {
		log.FromContext(ctx).InfoWithFields(log.Fields{"path": csvDumpPath}, "Skipping csv dump load")
		return
	}

	dir := filepath.Dir(csvDumpPath)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		panic(fmt.Errorf("failed to create dir for csv dump: %w", err))
	}

	var err error
	db.archivedCSVWithNamesDump, err = db.getCSVDumpFromDisk(ctx, csvDumpPath)
	if err != nil {
		panic(fmt.Errorf("failed to load GeoIP dump: %w", err))
	}
	log.FromContext(ctx).InfoWithFields(log.Fields{"path": csvDumpPath, "size MB": len(db.archivedCSVWithNamesDump) / 1024 / 1024}, "Dump loaded to memory")
}

func (db *maxmindDBWithCachedCSVDump) getCSVDumpFromDisk(ctx context.Context, dumpPath string) ([]byte, error) {
	path, err := db.Path()
	if err != nil {
		return nil, err
	}

	needUpdate, err := db.needUpdateDump(ctx, path, dumpPath)
	if err != nil {
		return nil, err
	}
	if !needUpdate {
		return os.ReadFile(dumpPath)
	}

	log.FromContext(ctx).InfoWithFields(log.Fields{"db": path, "csv": dumpPath}, "Updating CSV")
	return db.loadDumpFull(ctx, dumpPath)
}

func (db *maxmindDBWithCachedCSVDump) needUpdateDump(ctx context.Context, dbPath, dumpPath string) (bool, error) {
	dbStat, err := os.Stat(dbPath)
	if err != nil {
		return false, err
	}
	dumpStat, err := os.Stat(dumpPath)
	if err != nil {
		if os.IsNotExist(err) {
			return true, nil
		}
		return false, err
	}
	return dbStat.ModTime().After(dumpStat.ModTime()), nil
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
		return db.maxmindCSVDumper.WriteCSVTo(ctx, gw)
	}()
	if err != nil {
		return nil, os.Remove(temp)
	}

	return buf.Bytes(), os.Rename(temp, dumpPath)
}

func (db *maxmindDBWithCachedCSVDump) CSV(ctx context.Context, gzipCompress bool) (io.Reader, error) {
	select {
	case <-db.dumpReady:
		if !db.Available() {
			return nil, ErrDBNotAvailable
		}
		if db.archivedCSVWithNamesDump == nil {
			return nil, ErrGeoIPCSVDisabled
		}

		res := bytes.NewReader(db.archivedCSVWithNamesDump)
		if gzipCompress {
			return res, nil
		}
		return gzip.NewReader(res)
	default:
		return nil, ErrGeoIPCSVNotReady
	}
}
