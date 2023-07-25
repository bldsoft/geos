package repository

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/bldsoft/gost/log"
)

type maxmindDBWithCachedCSVDump struct {
	maxmindCSVDumper
	csvWithNamesDump []byte
	dumpReady        chan struct{}
}

func withCachedCSVDump(db maxmindCSVDumper) *maxmindDBWithCachedCSVDump {
	return &maxmindDBWithCachedCSVDump{maxmindCSVDumper: db, dumpReady: make(chan struct{})}
}

func (db *maxmindDBWithCachedCSVDump) initCSVDump(ctx context.Context, csvDumpPath string) {
	defer close(db.dumpReady)
	if filepath.Dir(csvDumpPath) == "." {
		return
	}

	if !db.Available() {
		log.FromContext(ctx).InfoWithFields(log.Fields{"path": csvDumpPath}, "Skipping csv dump load")
		return
	}

	dir := filepath.Dir(csvDumpPath)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		panic(fmt.Errorf("failed to create dir for csv dump: %w", err))
	}

	var err error
	db.csvWithNamesDump, err = db.getCSVDumpFromDisk(ctx, csvDumpPath)
	if err != nil {
		panic(fmt.Errorf("failed to load GeoIP dump: %w", err))
	}
	log.FromContext(ctx).InfoWithFields(log.Fields{"path": csvDumpPath, "size MB": len(db.csvWithNamesDump) / 1024 / 1024}, "Dump loaded to memory")
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
		return ioutil.ReadFile(dumpPath)
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

	if err := db.maxmindCSVDumper.WriteCSVTo(ctx, w); err != nil {
		return nil, os.Remove(temp)
	}
	return buf.Bytes(), os.Rename(temp, dumpPath)
}

func (db *maxmindDBWithCachedCSVDump) CSV(ctx context.Context, withColumnNames bool) ([]byte, error) {
	select {
	case <-db.dumpReady:
		if db.maxmindCSVDumper == nil {
			return nil, ErrDBNotAvailable
		}
		if db.csvWithNamesDump == nil {
			return nil, ErrGeoIPCSVDisabled
		}
		if !withColumnNames {
			return db.withoutColumnNames(), nil
		}
		return db.csvWithNamesDump, nil
	default:
		return nil, ErrGeoIPCSVNotReady
	}
}

func (db *maxmindDBWithCachedCSVDump) withoutColumnNames() []byte {
	i := bytes.Index(db.csvWithNamesDump, []byte("\n"))
	if i == -1 {
		return nil
	}
	return db.csvWithNamesDump[i+1:]
}
