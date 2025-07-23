package source

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/bldsoft/geos/pkg/entity"
	"github.com/mkrou/geonames"
	"golang.org/x/sync/errgroup"
)

const (
	geonamesBaseURL = "http://download.geonames.org/export/dump"
)

type GeoNamesSource struct {
	name entity.Subject

	CountriesFile      *UpdatableFile[ModTimeVersion]
	AdminDivisionsFile *UpdatableFile[ModTimeVersion]
	Cities500File      *UpdatableFile[ModTimeVersion]
}

func NewGeoNamesSource(dirPath string) *GeoNamesSource {
	res := &GeoNamesSource{
		name: entity.SubjectGeonames,
	}

	res.CountriesFile = NewTSUpdatableFile(
		fmt.Sprintf("%s/%s", dirPath, geonames.Countries.String()),
		fmt.Sprintf("%s/%s", geonamesBaseURL, geonames.Countries.String()),
	)
	res.AdminDivisionsFile = NewTSUpdatableFile(
		fmt.Sprintf("%s/%s", dirPath, geonames.AdminDivisions.String()),
		fmt.Sprintf("%s/%s", geonamesBaseURL, geonames.AdminDivisions.String()),
	)
	res.Cities500File = NewTSUpdatableFile(
		fmt.Sprintf("%s/%s", dirPath, string(geonames.Cities500)),
		fmt.Sprintf("%s/%s", geonamesBaseURL, string(geonames.Cities500)),
	)

	return res
}

func (s *GeoNamesSource) LastUpdateInterrupted(ctx context.Context) (bool, error) {
	for _, file := range []*UpdatableFile[ModTimeVersion]{
		s.CountriesFile,
		s.AdminDivisionsFile,
		s.Cities500File,
	} {
		interrupted, err := file.LastUpdateInterrupted(ctx)
		if err != nil {
			return false, err
		}
		if interrupted {
			return true, nil
		}
	}
	return false, nil
}

func (s *GeoNamesSource) Download(ctx context.Context) (entity.Updates, error) {
	var eg errgroup.Group
	updated := atomic.Bool{}

	for _, file := range []*UpdatableFile[ModTimeVersion]{
		s.CountriesFile,
		s.AdminDivisionsFile,
		s.Cities500File,
	} {
		eg.Go(func() error {
			u, err := file.Update(ctx)
			if err != nil {
				return nil
			}
			if u {
				updated.Store(true)
			}
			return nil
		})
	}

	upd := entity.Updates{}
	if err := eg.Wait(); err != nil {
		upd[s.name] = &entity.UpdateStatus{Error: err.Error()}
		return upd, nil
	}

	if updated.Load() {
		upd[s.name] = &entity.UpdateStatus{}
	}

	return upd, nil
}

func (s *GeoNamesSource) CheckUpdates(ctx context.Context) (entity.Updates, error) {
	var eg errgroup.Group
	var hasUpdates atomic.Bool

	for _, file := range []*UpdatableFile[ModTimeVersion]{
		s.CountriesFile,
		s.AdminDivisionsFile,
		s.Cities500File,
	} {
		eg.Go(func() error {
			available, err := file.CheckUpdates(ctx)
			if err != nil {
				return err
			}
			if available {
				hasUpdates.Store(true)
			}
			return nil
		})
	}

	upd := entity.Updates{}
	if err := eg.Wait(); err != nil {
		upd[s.name] = &entity.UpdateStatus{Error: err.Error()}
		return upd, nil
	}
	if hasUpdates.Load() {
		upd[s.name] = &entity.UpdateStatus{Available: true}
	}

	return upd, nil
}
