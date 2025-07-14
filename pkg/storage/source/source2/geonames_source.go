package source2

import (
	"context"
	"path/filepath"
	"sync/atomic"

	"github.com/bldsoft/geos/pkg/entity"
	"github.com/mkrou/geonames"
	"golang.org/x/sync/errgroup"
)

const (
	geonamesBaseURL = "http://download.geonames.org/export/dump"
)

var geonamesFiles = []string{
	geonames.Countries.String(),
	geonames.AdminDivisions.String(),
	string(geonames.Cities500),
}

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
		filepath.Join(dirPath, geonames.Countries.String()),
		filepath.Join(geonamesBaseURL, geonames.Countries.String()),
	)
	res.AdminDivisionsFile = NewTSUpdatableFile(
		filepath.Join(dirPath, geonames.AdminDivisions.String()),
		filepath.Join(geonamesBaseURL, geonames.AdminDivisions.String()),
	)
	res.Cities500File = NewTSUpdatableFile(
		filepath.Join(dirPath, string(geonames.Cities500)),
		filepath.Join(geonamesBaseURL, string(geonames.Cities500)),
	)

	return res
}

func (s *GeoNamesSource) Download(ctx context.Context) (entity.Updates, error) {
	var eg errgroup.Group

	for _, file := range []*UpdatableFile[ModTimeVersion]{
		s.CountriesFile,
		s.AdminDivisionsFile,
		s.Cities500File,
	} {
		eg.Go(func() error {
			return file.Update(ctx)
		})
	}

	if err := eg.Wait(); err != nil {
		return entity.Updates{
			s.name: &entity.UpdateStatus{
				Error: err.Error(),
			},
		}, nil
	}

	return entity.Updates{
		s.name: &entity.UpdateStatus{},
	}, nil
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

	if err := eg.Wait(); err != nil {
		return entity.Updates{
			s.name: &entity.UpdateStatus{
				Error: err.Error(),
			},
		}, nil
	}

	return entity.Updates{
		s.name: &entity.UpdateStatus{
			Available: hasUpdates.Load(),
		},
	}, nil
}

func (s *GeoNamesSource) DirPath() string {
	return filepath.Dir(s.CountriesFile.LocalPath)
}
