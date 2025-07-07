package source

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync/atomic"

	"github.com/bldsoft/geos/pkg/entity"
	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/utils/errgroup"
	"github.com/mkrou/geonames"
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
	dirPath  string
	name     entity.Subject
	managers map[string]*DownloadManager
}

func (s *GeoNamesSource) downloadFiles(ctx context.Context, filenames []string) error {
	if err := os.MkdirAll(s.dirPath, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", s.dirPath, err)
	}

	var eg errgroup.Group

	for _, filename := range filenames {
		eg.Go(func() error {
			manager := s.managers[filename]
			if manager == nil {
				return fmt.Errorf("download manager for %s not found", filename)
			}
			targetPath := filepath.Join(s.dirPath, filename)
			if err := manager.Download(ctx, fmt.Sprintf("%s/%s", geonamesBaseURL, filename), targetPath); err != nil {
				return fmt.Errorf("failed to download file %s: %w", filename, err)
			}
			return nil
		})
	}

	return eg.Wait()
}

func (s *GeoNamesSource) checkUpdates(ctx context.Context) (bool, error) {
	var eg errgroup.Group
	var hasUpdates int32

	for _, filename := range geonamesFiles {
		eg.Go(func() error {
			manager := s.managers[filename]
			if manager == nil {
				return fmt.Errorf("download manager for %s not found", filename)
			}

			targetURL := fmt.Sprintf("%s/%s", geonamesBaseURL, filename)
			targetPath := filepath.Join(s.dirPath, filename)

			hasUpdate, err := manager.CheckUpdates(ctx, targetURL, targetPath)
			if err != nil {
				return fmt.Errorf("failed to check updates for file %s: %w", filename, err)
			}

			if hasUpdate {
				atomic.StoreInt32(&hasUpdates, 1)
			}

			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return false, err
	}

	return atomic.LoadInt32(&hasUpdates) == 1, nil
}

func (s *GeoNamesSource) applyChanges(ctx context.Context) error {
	for _, filename := range geonamesFiles {
		manager := s.managers[filename]
		if manager == nil {
			continue
		}

		targetPath := filepath.Join(s.dirPath, filename)
		if err := manager.ApplyUpdate(ctx, targetPath); err != nil {
			continue
		}
	}

	return nil
}

func NewGeoNamesSource(dirPath string, autoUpdatePeriod int) *GeoNamesSource {
	s := &GeoNamesSource{
		dirPath:  dirPath,
		name:     entity.SubjectGeonames,
		managers: make(map[string]*DownloadManager),
	}

	ctx := context.Background()

	if len(dirPath) == 0 {
		log.FromContext(ctx).Warnf("Directory path for %s is not set. You will NOT be able to check for %s updates and download them without providing directory path.", s.name, s.name)
		return s
	}

	for _, filename := range geonamesFiles {
		s.managers[filename] = NewDownloadManager(s.name)
	}

	for file, manager := range s.managers {
		if err := manager.RecoverInterruptedDownloads(ctx, filepath.Join(s.dirPath, file), fmt.Sprintf("%s/%s", geonamesBaseURL, file)); err != nil {
			log.FromContext(ctx).ErrorfWithFields(log.Fields{"err": err}, "Failed to handle interrupted downloads for %s", s.name)
			return s
		}
	}

	// if err := s.initAutoUpdates(ctx, autoUpdatePeriod); err != nil {
	// 	log.FromContext(ctx).ErrorfWithFields(log.Fields{"err": err}, "Failed to initialize auto updates for %s", s.name)
	// }

	if autoUpdatePeriod == 0 {
		return s
	}

	if err := s.downloadFiles(ctx, geonamesFiles); err != nil {
		log.FromContext(ctx).ErrorfWithFields(log.Fields{"err": err}, "Failed to download missing %s files", s.name)
	}

	exist, err := s.checkUpdates(ctx)
	if err != nil {
		log.FromContext(ctx).ErrorfWithFields(log.Fields{"err": err}, "Failed to check for updates for %s", s.name)
		return s
	}

	if !exist {
		s.removeTmpFiles()
		return s
	}

	log.FromContext(ctx).Infof("Found updates for %s", s.name)
	if err := s.applyChanges(ctx); err != nil {
		log.FromContext(ctx).ErrorfWithFields(log.Fields{"err": err}, "Failed to apply updates for %s", s.name)
		return s
	}

	log.FromContext(ctx).Infof("Successfully applied updates for %s", s.name)

	return s
}

func (s *GeoNamesSource) DirPath() string {
	return s.dirPath
}

func (s *GeoNamesSource) removeTmpFiles() {
	for _, filename := range geonamesFiles {
		manager := s.managers[filename]
		if manager == nil {
			continue
		}

		manager.RemoveTmp(filename)
	}
}

// takes a lot of time for some reason
func (s *GeoNamesSource) CheckUpdates(ctx context.Context) (entity.Updates, error) {
	updates := entity.Updates{}

	if s.dirPath == "" {
		updates[s.name] = &entity.UpdateStatus{Error: fmt.Errorf("%s directory path is not set, unable to check for updates", s.name).Error()}
		return updates, nil
	}

	if err := s.downloadFiles(ctx, geonamesFiles); err != nil {
		updates[s.name] = &entity.UpdateStatus{Error: err.Error()}
		return updates, nil
	}

	exist, err := s.checkUpdates(ctx)
	if err != nil {
		updates[s.name] = &entity.UpdateStatus{Error: err.Error()}
		return updates, nil
	}

	if exist {
		updates[s.name] = &entity.UpdateStatus{Available: true}
	}

	s.removeTmpFiles()

	return updates, nil
}

func (s *GeoNamesSource) Download(ctx context.Context) (entity.Updates, error) {
	updates := entity.Updates{}

	if s.dirPath == "" {
		updates[s.name] = &entity.UpdateStatus{Error: fmt.Errorf("%s directory path is not set, unable to download", s.name).Error()}
		return updates, nil
	}

	if err := s.downloadFiles(ctx, geonamesFiles); err != nil {
		updates[s.name] = &entity.UpdateStatus{Error: err.Error()}
		return updates, nil
	}

	exist, err := s.checkUpdates(ctx)
	if err != nil {
		updates[s.name] = &entity.UpdateStatus{Error: err.Error()}
		return updates, nil
	}

	if !exist {
		return nil, nil
	}

	for _, manager := range s.managers {
		if err := manager.ApplyUpdate(ctx, filepath.Join(s.dirPath)); err != nil {
			updates[s.name] = &entity.UpdateStatus{Error: fmt.Errorf("failed to apply updates for %s: %w", s.name, err).Error()}
			return updates, nil
		}
	}

	log.FromContext(ctx).Infof("Applied updates for %s", s.name)

	updates[s.name] = &entity.UpdateStatus{}

	return updates, nil
}
