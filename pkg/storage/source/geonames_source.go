package source

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bldsoft/geos/pkg/entity"
	"github.com/bldsoft/gost/log"
)

const (
	geonamesBaseURL = "http://download.geonames.org/export/dump"
)

var geonamesFiles = []string{
	"countryInfo.txt",
	"admin1CodesASCII.txt",
	"cities500.zip",
}

type GeoNamesSource struct {
	dirPath string
	name    entity.Subject
}

func NewGeoNamesSource(dirPath string, autoUpdatePeriod int) *GeoNamesSource {
	s := &GeoNamesSource{
		dirPath: dirPath,
		name:    entity.SubjectGeonames,
	}

	ctx := context.Background()

	if len(dirPath) == 0 {
		log.FromContext(ctx).Warnf("Directory path for %s is not set. You will NOT be able to check for %s updates and download them without providing directory path.", s.name, s.name)
		return s
	}

	if err := s.initAutoUpdates(ctx, autoUpdatePeriod); err != nil {
		log.FromContext(ctx).ErrorfWithFields(log.Fields{"err": err}, "Failed to initialize auto updates for %s", s.name)
	}

	missing, err := s.checkMissingFiles()
	if err != nil {
		log.FromContext(ctx).ErrorfWithFields(log.Fields{"err": err}, "Failed to check existing files for %s", s.name)
		return s
	}

	if len(missing) > 0 {
		log.FromContext(ctx).Infof("Missing %s files: %s. Downloading...", s.name, strings.Join(missing, ", "))
		if err := s.downloadFiles(ctx, missing); err != nil {
			log.FromContext(ctx).ErrorfWithFields(log.Fields{"err": err}, "Failed to download missing %s files", s.name)
		}
	}

	if autoUpdatePeriod == 0 {
		return s
	}

	exist, err := s.checkUpdates(ctx)
	if err != nil {
		log.FromContext(ctx).ErrorfWithFields(log.Fields{"err": err}, "Failed to check for updates for %s", s.name)
		return s
	}

	if !exist {
		return s
	}

	log.FromContext(ctx).Infof("Found updates for %s during initialization", s.name)
	if err := s.downloadFiles(ctx, geonamesFiles); err != nil {
		log.FromContext(ctx).ErrorfWithFields(log.Fields{"err": err}, "Failed to download %s updates", s.name)
	}

	return s
}

func (s *GeoNamesSource) DirPath() string {
	return s.dirPath
}

func (s *GeoNamesSource) initAutoUpdates(ctx context.Context, autoUpdatePeriod int) error {
	if autoUpdatePeriod <= 0 {
		return nil
	}

	if s.dirPath == "" {
		return fmt.Errorf("missing required directory path")
	}

	go func() {
		timer := time.NewTicker(time.Duration(autoUpdatePeriod) * time.Hour)
		defer timer.Stop()

		for range timer.C {
			log.FromContext(ctx).Infof("Executing auto update for %s", s.name)

			exist, err := s.checkUpdates(ctx)
			if err != nil {
				log.FromContext(ctx).ErrorfWithFields(log.Fields{"err": err}, "Failed to check for updates for %s", s.name)
				continue
			}

			if !exist {
				log.FromContext(ctx).Infof("No updates found for %s", s.name)
				continue
			}

			log.FromContext(ctx).Infof("Found updates for %s", s.name)

			if err := s.downloadFiles(ctx, geonamesFiles); err != nil {
				log.FromContext(ctx).ErrorfWithFields(log.Fields{"err": err}, "Failed to download updates for %s", s.name)
				continue
			}

			log.FromContext(ctx).Infof("Successfully applied updates for %s", s.name)
		}
	}()

	return nil
}

func (s *GeoNamesSource) CheckUpdates(ctx context.Context) (entity.Updates, error) {
	updates := entity.Updates{}

	if s.dirPath == "" {
		updates[s.name] = &entity.UpdateStatus{Error: fmt.Errorf("%s directory path is not set, unable to check for updates", s.name).Error()}
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

	return updates, nil
}

func (s *GeoNamesSource) Download(ctx context.Context) (entity.Updates, error) {
	updates := entity.Updates{}

	if s.dirPath == "" {
		updates[s.name] = &entity.UpdateStatus{Error: fmt.Errorf("%s directory path is not set, unable to download", s.name).Error()}
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

	if err := s.downloadFiles(ctx, geonamesFiles); err != nil {
		updates[s.name] = &entity.UpdateStatus{Error: err.Error()}
		return updates, nil
	}

	log.FromContext(ctx).Infof("Applied updates for %s", s.name)

	updates[s.name] = &entity.UpdateStatus{}

	return updates, nil
}

func (s *GeoNamesSource) checkMissingFiles() ([]string, error) {
	var missing []string
	for _, filename := range geonamesFiles {
		fullPath := filepath.Join(s.dirPath, filename)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			missing = append(missing, filename)
		} else if err != nil {
			return nil, fmt.Errorf("failed to check file %s: %w", fullPath, err)
		}
	}
	return missing, nil
}

func (s *GeoNamesSource) checkUpdates(ctx context.Context) (bool, error) {
	for _, filename := range geonamesFiles {
		hasUpdate, err := s.checkFileUpdate(ctx, filename)
		if err != nil {
			return false, fmt.Errorf("failed to check updates for file %s: %w", filename, err)
		}
		if hasUpdate {
			return true, nil
		}
	}
	return false, nil
}

func (s *GeoNamesSource) checkFileUpdate(ctx context.Context, filename string) (bool, error) {
	url := fmt.Sprintf("%s/%s", geonamesBaseURL, filename)

	req, err := http.NewRequestWithContext(ctx, "HEAD", url, nil)
	if err != nil {
		return false, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("failed to check file %s: %s", filename, resp.Status)
	}

	remoteLastModified := resp.Header.Get("Last-Modified")
	if remoteLastModified == "" {
		return true, nil
	}

	remoteTime, err := time.Parse(time.RFC1123, remoteLastModified)
	if err != nil {
		return true, nil
	}

	localPath := filepath.Join(s.dirPath, filename)
	stat, err := os.Stat(localPath)
	if os.IsNotExist(err) {
		return true, nil
	} else if err != nil {
		return false, err
	}

	return remoteTime.After(stat.ModTime().Add(-time.Minute)), nil
}

func (s *GeoNamesSource) downloadFiles(ctx context.Context, filenames []string) error {
	if err := os.MkdirAll(s.dirPath, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", s.dirPath, err)
	}

	for _, filename := range filenames {
		if err := s.downloadFile(ctx, filename); err != nil {
			return fmt.Errorf("failed to download file %s: %w", filename, err)
		}
	}
	return nil
}

func (s *GeoNamesSource) downloadFile(ctx context.Context, filename string) error {
	url := fmt.Sprintf("%s/%s", geonamesBaseURL, filename)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download file: %s", resp.Status)
	}

	tmpPath := filepath.Join(s.dirPath, filename+".tmp")
	tmpFile, err := os.Create(tmpPath)
	if err != nil {
		return err
	}
	defer os.Remove(tmpPath)

	_, err = io.Copy(tmpFile, resp.Body)
	tmpFile.Close()
	if err != nil {
		return err
	}

	finalPath := filepath.Join(s.dirPath, filename)
	if err := os.Rename(tmpPath, finalPath); err != nil {
		return fmt.Errorf("failed to move temporary file to %s: %w", finalPath, err)
	}

	return nil
}
