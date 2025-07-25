package repository

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/bldsoft/geos/pkg/entity"
	"github.com/bldsoft/geos/pkg/storage/maxmind"
	"github.com/bldsoft/geos/pkg/storage/source"
	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/version"
	"github.com/oschwald/maxminddb-golang"
)

type dumpMetaData struct {
	*maxminddb.Metadata
	BuildVersion string `json:"BuildVersion"`
}

type maxmindDBWithCachedCSVDump struct {
	maxmind.CSVDumper
	archivedCSVWithNamesDump atomic.Pointer[[]byte]
	csvDumpPath              string
	fileRepository           source.FileRepository
}

func withCachedCSVDump[T maxmind.CSVEntity](db maxmind.Database, csvDumpPath string) *maxmindDBWithCachedCSVDump {
	res := &maxmindDBWithCachedCSVDump{
		CSVDumper:      maxmind.NewCSVDumper[T](db),
		csvDumpPath:    csvDumpPath + ".gz",
		fileRepository: source.NewLocalFileRepository(),
	}
	res.initCSVDump(context.Background())
	return res
}

func (db *maxmindDBWithCachedCSVDump) initCSVDump(ctx context.Context) {
	logger := log.FromContext(ctx).WithFields(log.Fields{"db type": db.metadata().DatabaseType})
	ctx = context.WithValue(ctx, log.LoggerCtxKey, logger)

	r, err := db.fileRepository.Reader(ctx, db.csvDumpPath)
	if err != nil {
		logger.InfoWithFields(log.Fields{"err": err}, "Failed to open GeoIP dump")
		return
	}
	defer r.Close()

	data, err := io.ReadAll(r)
	if err != nil {
		logger.InfoWithFields(log.Fields{"err": err}, "Failed to read GeoIP dump")
	}
	db.archivedCSVWithNamesDump.Store(&data)
}

func (db *maxmindDBWithCachedCSVDump) CSV(ctx context.Context, gzipCompress bool) (io.Reader, error) {
	if data := db.archivedCSVWithNamesDump.Load(); data != nil {
		res := bytes.NewReader(*data)
		if gzipCompress {
			return res, nil
		}
		return gzip.NewReader(res)
	}
	return nil, ErrGeoIPCSVNotReady
}

func (db *maxmindDBWithCachedCSVDump) Update(ctx context.Context, force bool) error {
	if err := db.CSVDumper.Update(ctx, force); err != nil {
		return err
	}

	if err := db.updateDumpIfNeeded(ctx); err != nil {
		return err
	}
	return nil
}

func (db *maxmindDBWithCachedCSVDump) updateDumpIfNeeded(ctx context.Context) error {
	if filepath.Dir(db.csvDumpPath) == "." {
		return nil
	}

	if needUpdate, err := db.needUpdateDump(ctx); err != nil || !needUpdate {
		return err
	}

	err := db.updateDump(ctx)
	if err != nil {
		return fmt.Errorf("failed to load GeoIP dump: %w", err)
	}
	return nil
}

func (db *maxmindDBWithCachedCSVDump) needUpdateDump(ctx context.Context) (bool, error) {
	exists, err := db.fileRepository.Exists(ctx, db.csvDumpPath)
	if err != nil {
		return false, err
	}
	if !exists {
		return true, nil
	}

	dumpMetaData, err := db.dumpMetaData()
	if err != nil {
		log.FromContext(ctx).DebugWithFields(log.Fields{"err": err}, "Failed to get dump metadata")
		return true, nil
	}

	if dumpMetaData.BuildVersion != version.Version {
		return true, nil
	}

	dumpDBBuildTime := time.Unix(int64(dumpMetaData.BuildEpoch), 0)
	dbBuildTime := time.Unix(int64(db.metadata().BuildEpoch), 0)
	return dbBuildTime.After(dumpDBBuildTime), nil
}

func (db *maxmindDBWithCachedCSVDump) updateDump(ctx context.Context) error {
	log.FromContext(ctx).InfoWithFields(log.Fields{"meta": db.metadata(), "csv": db.csvDumpPath}, "Updating CSV")

	temp := db.tempDumpPath()

	var buf bytes.Buffer
	tmpFile, err := db.fileRepository.CreateIfNotExists(ctx, temp)
	if err != nil {
		if errors.Is(err, source.ErrFileExists) {
			// already updating
			return nil
		}
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer db.fileRepository.Remove(ctx, temp)
	defer tmpFile.Close()

	w := io.MultiWriter(&buf, tmpFile)

	err = func() error {
		gw := gzip.NewWriter(w)
		defer gw.Close()
		return db.CSVDumper.WriteCSVTo(ctx, gw)
	}()
	if err != nil {
		return err
	}

	if err := db.writeDumpMetadata(); err != nil {
		return err
	}

	err = db.fileRepository.Rename(ctx, temp, db.csvDumpPath)
	if err != nil {
		return err
	}

	data := buf.Bytes()
	db.archivedCSVWithNamesDump.Store(&data)

	log.FromContext(ctx).InfoWithFields(log.Fields{
		"path":    db.csvDumpPath,
		"size MB": len(data) / 1024 / 1024,
	}, "Dump loaded to memory")

	return nil
}

func (db *maxmindDBWithCachedCSVDump) tempDumpPath() string {
	return db.csvDumpPath + ".tmp"
}

func (db *maxmindDBWithCachedCSVDump) writeDumpMetadata() error {
	data, err := json.Marshal(db.metadata())
	if err != nil {
		return err
	}
	return db.fileRepository.Write(context.Background(), db.dumpMetaDataPath(), bytes.NewReader(data))
}

func (db *maxmindDBWithCachedCSVDump) dumpMetaData() (*dumpMetaData, error) {
	dumpMetaDataPath := db.dumpMetaDataPath()

	r, err := db.fileRepository.Reader(context.Background(), dumpMetaDataPath)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	var meta dumpMetaData
	if err := json.NewDecoder(r).Decode(&meta); err != nil {
		return nil, err
	}

	return &meta, nil
}

func (db *maxmindDBWithCachedCSVDump) metadata() *dumpMetaData {
	meta, _ := db.MetaData()
	return &dumpMetaData{Metadata: meta, BuildVersion: version.Version}
}

func (db *maxmindDBWithCachedCSVDump) dumpMetaDataPath() string {
	ext := filepath.Ext(db.csvDumpPath)
	return strings.TrimSuffix(db.csvDumpPath, ext) + ".meta"
}

func (db *maxmindDBWithCachedCSVDump) CheckUpdates(ctx context.Context) (entity.Update, error) {
	return db.CSVDumper.CheckUpdates(ctx)
}
