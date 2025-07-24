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
	CountriesFile      *UpdatableFile[ModTimeVersion]
	AdminDivisionsFile *UpdatableFile[ModTimeVersion]
	Cities500File      *UpdatableFile[ModTimeVersion]
}

func NewGeoNamesSource(dirPath string) *GeoNamesSource {
	res := &GeoNamesSource{}

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

func (s *GeoNamesSource) TryUpdate(ctx context.Context) error {
	var eg errgroup.Group

	for _, file := range []*UpdatableFile[ModTimeVersion]{
		s.CountriesFile,
		s.AdminDivisionsFile,
		s.Cities500File,
	} {
		eg.Go(func() error {
			return file.TryUpdate(ctx)
		})
	}

	return eg.Wait()
}

func (s *GeoNamesSource) CheckUpdates(ctx context.Context) (entity.Update, error) {
	var eg errgroup.Group
	var res atomic.Pointer[entity.Update]

	for _, file := range []*UpdatableFile[ModTimeVersion]{
		s.CountriesFile,
		s.AdminDivisionsFile,
		s.Cities500File,
	} {
		eg.Go(func() error {
			update, err := file.CheckUpdates(ctx)
			if err != nil {
				return err
			}
			if res.Load() == nil || update.AvailableVersion != "" {
				res.Store(&update)
			}
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return entity.Update{}, err
	}

	return *res.Load(), nil
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
