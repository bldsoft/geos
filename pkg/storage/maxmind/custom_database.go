package maxmind

import (
	"context"

	"github.com/bldsoft/geos/pkg/entity"
	"github.com/bldsoft/geos/pkg/storage/source"
	"github.com/bldsoft/geos/pkg/storage/state"
	"github.com/bldsoft/gost/log"
	"github.com/oschwald/maxminddb-golang"
)

type CustomDatabase struct {
	*MultiMaxMindDB[*DatabasePatch]
	source *source.PatchesSource
}

func NewCustomDatabaseFromTarGz(source *source.PatchesSource) *CustomDatabase {
	customDBs, err := NewDatabasePatchesFromTarGz(source)
	if err != nil {
		log.Logger.Errorf("failed to load custom databases %s: %v", source.Name, err)
	}
	res := &CustomDatabase{
		MultiMaxMindDB: NewMultiMaxMindDB(customDBs...),
		source:         source,
	}
	return res
}

func (db *CustomDatabase) MetaData() (*maxminddb.Metadata, error) {
	if len(db.dbs) == 0 {
		return &maxminddb.Metadata{}, nil //there is no problem if there are no patches
	}
	return db.MultiMaxMindDB.MetaData()
}

func (db *CustomDatabase) Download(ctx context.Context) (updates entity.Updates, err error) {
	if updates, err = db.source.Download(ctx); err != nil {
		return nil, err
	}

	customDBs, err := NewDatabasePatchesFromTarGz(db.source)
	if err != nil {
		return nil, err
	}

	db.MultiMaxMindDB = NewMultiMaxMindDB(customDBs...)
	return
}

func (db *CustomDatabase) CheckUpdates(ctx context.Context) (entity.Updates, error) {
	return db.source.CheckUpdates(ctx)
}

func (db *CustomDatabase) LastUpdateInterrupted(ctx context.Context) (bool, error) {
	return db.source.LastUpdateInterrupted(ctx)
}

func (db *CustomDatabase) State() *state.GeosState {
	ctx := context.TODO()
	result := &state.GeosState{}

	var archiveTimestamp int64
	if v, err := db.source.Version(ctx); err == nil {
		archiveTimestamp = v.Time().Unix()
	} else {
		var maxTimestamp int64
		for _, patch := range db.dbs {
			if patch.state > maxTimestamp {
				maxTimestamp = patch.state
			}
		}
		archiveTimestamp = maxTimestamp
	}

	if archiveTimestamp > 0 {
		switch db.source.Name {
		case entity.SubjectCitiesDbPatches:
			result.CityPatchesTimestamp = archiveTimestamp
		case entity.SubjectISPDbPatches:
			result.ISPPatchesTimestamp = archiveTimestamp
		}
	}

	return result
}
