package source

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/bldsoft/geos/pkg/entity"
	"github.com/bldsoft/gost/log"
)

type PatchesSource struct {
	sourceUrl string
	dirPath   string
	prefix    string
	Name      entity.Subject
}

func (s *PatchesSource) tmpFilePath() string {
	return fmt.Sprintf("%s/%s_patches_tmp.tar.gz", s.dirPath, s.prefix)
}

func (s *PatchesSource) ArchiveFilePath() string {
	return fmt.Sprintf("%s/%s_patches.tar.gz", s.dirPath, s.prefix)
}

func NewPatchesSource(sourceUrl, dirPath, prefix string, name entity.Subject, autoUpdatePeriod int) *PatchesSource {
	s := &PatchesSource{
		sourceUrl: sourceUrl,
		dirPath:   dirPath,
		prefix:    prefix,
		Name:      name,
	}

	ctx := context.Background()

	if len(sourceUrl) == 0 {
		log.FromContext(ctx).Warnf("Source for %s is not set. You will NOT be able to check for %s updates and download them without providing source.", s.Name, s.Name)
		return s
	}

	if err := s.initAutoUpdates(ctx, autoUpdatePeriod); err != nil {
		log.FromContext(ctx).ErrorfWithFields(log.Fields{"err": err}, "Failed to initialize auto updates for %s", s.Name)
	}

	_, err := os.Stat(s.ArchiveFilePath())
	if err != nil {
		if !os.IsNotExist(err) {
			log.FromContext(ctx).Errorf("Failed to check if %s archive exists: %v", s.Name, err)
			return s
		}

		log.FromContext(ctx).Warnf("No %s archive found, creating empty one", s.Name)
		if _, err := os.Create(s.ArchiveFilePath()); err != nil {
			log.FromContext(ctx).Errorf("Failed to create empty archive for %s: %v", s.Name, err) //TEST THIS
		}
	}

	if autoUpdatePeriod == 0 {
		return s
	}

	if err := s.getTmpArchive(s.sourceUrl); err != nil {
		log.FromContext(ctx).Errorf("Failed to check for updates for %s: %v", s.Name, err)
		return s
	}

	exist, err := s.checkUpdates()
	if err != nil {
		log.FromContext(ctx).Errorf("Failed to check for updates for %s: %v", s.Name, err)
		return s
	}

	if !exist {
		if err := s.rmTmpArchive(); err != nil {
			log.FromContext(ctx).Errorf("Failed to remove temporary archive for %s: %v", s.Name, err)
		}
		return s
	}

	log.FromContext(ctx).Infof("Found updates for %s", s.Name)
	if err := s.updateLocalFiles(); err != nil {
		log.FromContext(ctx).Errorf("Failed to apply updates for %s: %v", s.Name, err)
		return s
	}

	log.FromContext(ctx).Infof("Successfully applied updates for %s", s.Name)

	return s
}

func (s *PatchesSource) initAutoUpdates(ctx context.Context, autoUpdatePeriod int) error {
	if autoUpdatePeriod <= 0 {
		return nil
	}

	if s.sourceUrl == "" || s.dirPath == "" {
		return fmt.Errorf("missing required params")
	}

	go func() {
		timer := time.NewTicker(time.Duration(autoUpdatePeriod) * time.Hour)
		defer timer.Stop()

		for range timer.C {
			log.FromContext(ctx).Infof("Executing auto update for %s", s.Name)

			err := s.getTmpArchive(s.sourceUrl)
			if err != nil {
				log.FromContext(ctx).ErrorfWithFields(log.Fields{"err": err}, "failed to get source files for %s", s.Name)
				continue
			}

			exist, err := s.checkUpdates()
			if err != nil {
				log.FromContext(ctx).ErrorfWithFields(log.Fields{"err": err}, "failed to check for updates for %s", s.Name)
				continue
			}

			if !exist {
				log.FromContext(ctx).Infof("No updates found for %s", s.Name)
				continue
			}

			log.FromContext(ctx).Infof("Found update for %s", s.Name)

			if err := s.updateLocalFiles(); err != nil {
				log.FromContext(ctx).ErrorfWithFields(log.Fields{"err": err}, "failed to download updates for %s", s.Name)
				continue
			}

			log.FromContext(ctx).Infof("Successfully applied updates for %s", s.Name)
		}
	}()

	return nil
}

func (s *PatchesSource) CheckUpdates(ctx context.Context) (entity.Updates, error) {
	updates := entity.Updates{}

	if s.dirPath == "" {
		updates[s.Name] = &entity.UpdateStatus{Error: fmt.Errorf("%s database path is not set, unable to check for updates", s.Name).Error()}
		return updates, nil
	}

	if s.sourceUrl == "" {
		updates[s.Name] = &entity.UpdateStatus{Error: fmt.Errorf("%s source is not set, unable to check for updates", s.Name).Error()}
		return updates, nil
	}

	err := s.getTmpArchive(s.sourceUrl)
	if err != nil {
		updates[s.Name] = &entity.UpdateStatus{Error: err.Error()}
		return updates, nil
	}

	exist, err := s.checkUpdates()
	if err != nil {
		updates[s.Name] = &entity.UpdateStatus{Error: err.Error()}
		return updates, nil
	}

	if exist {
		updates[s.Name] = &entity.UpdateStatus{Available: true}
	}

	if err := s.rmTmpArchive(); err != nil {
		return nil, fmt.Errorf("failed to remove temporary archive: %w", err)
	}

	return updates, nil
}

func (s *PatchesSource) checkUpdates() (bool, error) {
	tmpFile, err := os.Open(s.tmpFilePath())
	if err != nil {
		return false, err
	}
	defer tmpFile.Close()

	currentFile, err := os.Open(s.ArchiveFilePath())
	if err != nil {
		return false, err
	}
	defer currentFile.Close()

	tmpContent, err := io.ReadAll(tmpFile)
	if err != nil {
		return false, err
	}

	currentContent, err := io.ReadAll(currentFile)
	if err != nil {
		return false, err
	}

	return !bytes.Equal(tmpContent, currentContent), nil
}

func (s *PatchesSource) Download(ctx context.Context) (entity.Updates, error) {
	updates := entity.Updates{}

	if s.dirPath == "" {
		updates[s.Name] = &entity.UpdateStatus{Error: fmt.Errorf("%s path is not set, unable to check for updates", s.Name).Error()}
		return updates, nil
	}

	if s.sourceUrl == "" {
		updates[s.Name] = &entity.UpdateStatus{Error: fmt.Errorf("%s source is not set, unable to check for updates", s.Name).Error()}
		return updates, nil
	}

	err := s.getTmpArchive(s.sourceUrl)
	if err != nil {
		updates[s.Name] = &entity.UpdateStatus{Error: err.Error()}
		return updates, nil
	}

	exist, err := s.checkUpdates()
	if err != nil {
		updates[s.Name] = &entity.UpdateStatus{Error: err.Error()}
		return updates, nil
	}

	if !exist {
		return nil, nil
	}

	if err := s.updateLocalFiles(); err != nil {
		updates[s.Name] = &entity.UpdateStatus{Error: err.Error()}
		return updates, err
	}

	log.FromContext(ctx).Infof("Applied updates for %s", s.Name)

	updates[s.Name] = &entity.UpdateStatus{}

	return updates, nil
}

func (s *PatchesSource) rmTmpArchive() error {
	return os.Remove(s.tmpFilePath())
}

func (s *PatchesSource) updateLocalFiles() error {
	if err := os.Remove(s.ArchiveFilePath()); err != nil && !os.IsNotExist(err) {
		return err
	}

	return os.Rename(s.tmpFilePath(), s.ArchiveFilePath())
}

func (s *PatchesSource) DirPath() string {
	return s.dirPath
}

func (s *PatchesSource) getTmpArchive(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad response status: %s", resp.Status)
	}

	tmpFile, err := os.Create(s.tmpFilePath())
	if err != nil {
		return err
	}
	defer tmpFile.Close()

	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		return err
	}

	if err := tmpFile.Close(); err != nil {
		return err
	}

	return nil
}
