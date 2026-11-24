package asn

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/bldsoft/geos/pkg/entity"
	"github.com/bldsoft/geos/pkg/storage/maxmind"
	"github.com/bldsoft/geos/pkg/utils"
	"github.com/bldsoft/gost/log"
)

var ErrASNNotReady = fmt.Errorf("asn catalog is %w", utils.ErrNotReady)

type ASNStorage struct {
	db      maxmind.Database
	catalog atomic.Pointer[index]
}

func NewASNStorage(ctx context.Context, db maxmind.Database, syncInit ...bool) *ASNStorage {
	s := &ASNStorage{db: db}
	if len(syncInit) > 0 && syncInit[0] {
		s.fill(ctx)
	} else {
		go s.fill(ctx)
	}
	return s
}

func (s *ASNStorage) fill(ctx context.Context) {
	items, err := collectASNs(ctx, s.db)
	if err != nil {
		log.FromContext(ctx).ErrorWithFields(log.Fields{"err": err}, "Failed to build ASN catalog")
		return
	}
	idx := &index{}
	idx.Init(items)
	s.catalog.Store(idx)
	log.FromContext(ctx).Infof("ASN catalog: %d", len(items))
}

func (s *ASNStorage) Update(ctx context.Context) {
	s.fill(ctx)
}

func (s *ASNStorage) GetFiltered(ctx context.Context, filter entity.ASNFilter) ([]*entity.ASN, error) {
	idx := s.catalog.Load()
	if idx == nil {
		return nil, ErrASNNotReady
	}
	filtered := idx.GetFiltered(filter)
	if filter.Limit != 0 && len(filtered) > int(filter.Limit) {
		return filtered[:filter.Limit], nil
	}
	return filtered, nil
}

func collectASNs(ctx context.Context, db maxmind.Database) ([]*entity.ASN, error) {
	networks, err := db.Networks(ctx)
	if err != nil {
		return nil, err
	}
	seen := make(map[uint32]string)
	for result := range networks {
		if err := result.Err(); err != nil {
			return nil, err
		}
		var rec entity.ISP
		if err := result.Decode(&rec); err != nil {
			return nil, err
		}
		if rec.AutonomousSystemNumber == 0 {
			continue
		}
		if _, ok := seen[rec.AutonomousSystemNumber]; !ok {
			seen[rec.AutonomousSystemNumber] = rec.AutonomousSystemOrganization
		}
	}
	items := make([]*entity.ASN, 0, len(seen))
	for number, org := range seen {
		items = append(items, &entity.ASN{
			ASN: number,
			ASO: org,
		})
	}
	return items, nil
}
