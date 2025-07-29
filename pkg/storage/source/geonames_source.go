package source

import (
	"context"
	"net/url"
	"path/filepath"
	"sync/atomic"

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

	join := func(base, path string) string {
		url, err := url.JoinPath(base, path)
		if err != nil {
			panic(err)
		}
		return url
	}
	res.CountriesFile = NewTSUpdatableFile(
		filepath.Join(dirPath, geonames.Countries.String()),
		join(geonamesBaseURL, geonames.Countries.String()),
	)
	res.AdminDivisionsFile = NewTSUpdatableFile(
		filepath.Join(dirPath, geonames.AdminDivisions.String()),
		join(geonamesBaseURL, geonames.AdminDivisions.String()),
	)
	res.Cities500File = NewTSUpdatableFile(
		filepath.Join(dirPath, string(geonames.Cities500)),
		join(geonamesBaseURL, string(geonames.Cities500)),
	)

	return res
}

func (s *GeoNamesSource) Update(ctx context.Context, force bool) error {
	var eg errgroup.Group

	for _, file := range []*UpdatableFile[ModTimeVersion]{
		s.CountriesFile,
		s.AdminDivisionsFile,
		s.Cities500File,
	} {
		eg.Go(func() error {
			return file.Update(ctx, force)
		})
	}

	return eg.Wait()
}

func (s *GeoNamesSource) CheckUpdates(ctx context.Context) (Update[ModTimeVersion], error) {
	var eg errgroup.Group
	var res atomic.Pointer[Update[ModTimeVersion]]

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
			if res.Load() == nil || update.RemoteVersion.IsHigher(res.Load().RemoteVersion) {
				res.Store(&update)
			}
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return Update[ModTimeVersion]{}, err
	}

	return *res.Load(), nil
}
