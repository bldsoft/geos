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

	"github.com/bldsoft/geos/pkg/entity"
	"github.com/bldsoft/geos/pkg/storage/maxmind"
	"github.com/bldsoft/geos/pkg/storage/source"
	"github.com/bldsoft/geos/pkg/utils"
	"github.com/bldsoft/gost/log"
)

type maxmindDBWithCachedCSVDump struct {
	*maxmind.PatchedDatabase
	csvDumper                maxmind.CSVDumper
	archivedCSVWithNamesDump atomic.Pointer[[]byte]
	csvDumpPath              string
	fileRepository           source.FileRepository
}

func withCachedCSVDump[T maxmind.CSVEntity](
	ctx context.Context,
	db *maxmind.PatchedDatabase,
	csvDumpPath string,
) *maxmindDBWithCachedCSVDump {
	res := &maxmindDBWithCachedCSVDump{
		PatchedDatabase: db,
		csvDumper:       maxmind.NewCSVDumper[T](db),
		csvDumpPath:     csvDumpPath + ".gz",
		fileRepository:  source.NewLocalFileRepository(),
	}
	res.initCSVDump(ctx)
	return res
}

func (db *maxmindDBWithCachedCSVDump) initCSVDump(ctx context.Context) {
	r, err := db.fileRepository.Reader(ctx, db.csvDumpPath)
	if err == nil {
		defer r.Close()
		data, err := io.ReadAll(r)
		if err != nil {
			log.FromContext(ctx).InfoWithFields(log.Fields{"err": err}, "Failed to read GeoIP dump")
			return
		}
		db.archivedCSVWithNamesDump.Store(&data)
		return
	}

	log.FromContext(ctx).InfoWithFields(log.Fields{"err": err}, "Failed to open GeoIP dump")
	go func() {
		if err := db.updateDump(ctx, true); err != nil {
			log.FromContext(ctx).InfoWithFields(log.Fields{"err": err}, "Failed to update GeoIP dump")
		}
	}()
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

func (db *maxmindDBWithCachedCSVDump) WriteCSVTo(ctx context.Context, w io.Writer) error {
	if data := db.archivedCSVWithNamesDump.Load(); data != nil {
		_, err := w.Write(*data)
		return err
	}
	return ErrGeoIPCSVNotReady
}

func (db *maxmindDBWithCachedCSVDump) Update(ctx context.Context, force bool) error {
	updates, err := db.PatchedDatabase.CheckUpdates(ctx)
	if err != nil {
		return err
	}

	if updates.RemoteVersion.Compare(updates.CurrentVersion) > 0 {
		if err := db.PatchedDatabase.Update(ctx, force); err != nil {
			return err
		}
	}

	return db.updateDumpIfNeeded(ctx, force)
}

func (db *maxmindDBWithCachedCSVDump) updateDumpIfNeeded(ctx context.Context, force bool) error {
	if filepath.Dir(db.csvDumpPath) == "." {
		return nil
	}

	if needUpdate, err := db.needUpdateDump(ctx); err != nil || !needUpdate {
		return err
	}

	err := db.updateDump(ctx, force)
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

	dumpVersion, err := db.dumpVersion(ctx)
	if err != nil {
		log.FromContext(ctx).DebugWithFields(log.Fields{"err": err}, "Failed to get dump metadata")
		return true, nil
	}

	dbVersion, err := db.dbVersion(ctx)
	if err != nil {
		log.FromContext(ctx).ErrorWithFields(log.Fields{"err": err}, "Failed to get db metadata")
		return false, err
	}

	return dbVersion.Compare(dumpVersion) > 0, nil
}

func (db *maxmindDBWithCachedCSVDump) updateDump(ctx context.Context, force bool) error {
	log.FromContext(ctx).InfoWithFields(log.Fields{"csv": db.csvDumpPath}, "Updating CSV")

	temp := db.tempDumpPath()
	if force {
		_ = db.fileRepository.Remove(ctx, temp)
	}

	var buf bytes.Buffer
	tmpFile, err := db.fileRepository.CreateIfNotExists(ctx, temp)
	if err != nil {
		if errors.Is(err, source.ErrFileExists) {
			return utils.ErrUpdateInProgress
		}
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer db.fileRepository.Remove(ctx, temp)
	defer tmpFile.Close()

	w := io.MultiWriter(&buf, tmpFile)

	err = func() error {
		gw := gzip.NewWriter(w)
		defer gw.Close()
		return db.csvDumper.WriteCSVTo(ctx, gw)
	}()
	if err != nil {
		return err
	}

	err = db.fileRepository.Rename(ctx, temp, db.csvDumpPath)
	if err != nil {
		return err
	}

	data := buf.Bytes()
	db.archivedCSVWithNamesDump.Store(&data)

	version, err := db.dbVersion(ctx)
	if err != nil {
		return err
	}
	if err := db.writeDumpVersion(ctx, version); err != nil {
		return err
	}

	log.FromContext(ctx).InfoWithFields(log.Fields{
		"path":    db.csvDumpPath,
		"size MB": len(data) / 1024 / 1024,
	}, "Dump loaded to memory")

	return nil
}

func (db *maxmindDBWithCachedCSVDump) tempDumpPath() string {
	return db.csvDumpPath + ".tmp"
}

func (db *maxmindDBWithCachedCSVDump) writeDumpVersion(ctx context.Context, version entity.PatchedMMDBVersion) error {
	data, err := json.Marshal(version)
	if err != nil {
		return err
	}
	return db.fileRepository.Write(ctx, db.dumpMetaDataPath(), bytes.NewReader(data))
}

func (db *maxmindDBWithCachedCSVDump) dumpVersion(ctx context.Context) (entity.PatchedMMDBVersion, error) {
	dumpMetaDataPath := db.dumpMetaDataPath()

	r, err := db.fileRepository.Reader(ctx, dumpMetaDataPath)
	if err != nil {
		return entity.PatchedMMDBVersion{}, err
	}
	defer r.Close()

	var meta entity.PatchedMMDBVersion
	if err := json.NewDecoder(r).Decode(&meta); err != nil {
		return entity.PatchedMMDBVersion{}, err
	}

	// TODO:backward compatibility, remove in the future
	if meta.DB.Version == nil {
		version, err := db.dbVersion(ctx)
		if err != nil {
			return entity.PatchedMMDBVersion{}, err
		}
		err = db.writeDumpVersion(ctx, version)
		if err != nil {
			return entity.PatchedMMDBVersion{}, err
		}
		return version, nil
	}
	// ==============================

	return meta, nil
}

func (db *maxmindDBWithCachedCSVDump) dbVersion(ctx context.Context) (entity.PatchedMMDBVersion, error) {
	udpate, err := db.PatchedDatabase.CheckUpdates(ctx)
	if err != nil {
		return entity.PatchedMMDBVersion{}, err
	}
	return udpate.CurrentVersion, nil
}

func (db *maxmindDBWithCachedCSVDump) dumpMetaDataPath() string {
	ext := filepath.Ext(db.csvDumpPath)
	return strings.TrimSuffix(db.csvDumpPath, ext) + ".meta"
}

func (db *maxmindDBWithCachedCSVDump) CheckUpdates(ctx context.Context) (entity.Update[entity.PatchedMMDBVersion], error) {
	updates, err := db.PatchedDatabase.CheckUpdates(ctx)
	if err != nil {
		return entity.Update[entity.PatchedMMDBVersion]{}, err
	}

	if db.archivedCSVWithNamesDump.Load() != nil {
		dumpVersion, err := db.dumpVersion(ctx)
		if err != nil {
			return entity.Update[entity.PatchedMMDBVersion]{}, err
		}
		updates.CurrentVersion = dumpVersion
	}
	return updates, nil
}
