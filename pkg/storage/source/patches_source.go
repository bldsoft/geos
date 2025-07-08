package source

import (
	"context"
	"fmt"
	"os"

	"github.com/bldsoft/geos/pkg/entity"
	"github.com/bldsoft/gost/log"
)

type PatchesSource struct {
	sourceUrl       string
	dirPath         string
	prefix          string
	Name            entity.Subject
	downloadManager *DownloadManager
}

func (s *PatchesSource) ArchiveFilePath() string {
	return fmt.Sprintf("%s/%s_patches.tar.gz", s.dirPath, s.prefix)
}

func NewPatchesSource(sourceUrl, dirPath, prefix string, name entity.Subject, autoUpdatePeriod int) *PatchesSource {
	s := &PatchesSource{
		sourceUrl:       sourceUrl,
		dirPath:         dirPath,
		prefix:          prefix,
		Name:            name,
		downloadManager: NewDownloadManager(name),
	}

	ctx := context.Background()

	if len(sourceUrl) == 0 {
		log.FromContext(ctx).Warnf("Source for %s is not set. You will NOT be able to check for %s updates and download them without providing source.", s.Name, s.Name)
		return s
	}

	if err := s.downloadManager.RecoverInterruptedDownloads(ctx, s.ArchiveFilePath(), s.sourceUrl); err != nil {
		log.FromContext(ctx).ErrorfWithFields(log.Fields{"err": err}, "failed to handle interrupted downloads for %s", s.Name)
	}

	_, err := os.Stat(s.ArchiveFilePath())
	if err != nil {
		if !os.IsNotExist(err) {
			log.FromContext(ctx).ErrorfWithFields(log.Fields{"err": err}, "failed to check if %s archive exists", s.Name)
			return s
		}

		log.FromContext(ctx).Warnf("No %s archive found, creating empty one", s.Name)
		if _, err := os.Create(s.ArchiveFilePath()); err != nil {
			log.FromContext(ctx).ErrorfWithFields(log.Fields{"err": err}, "failed to create empty archive for %s", s.Name)
		}
	}

	if autoUpdatePeriod == 0 {
		return s
	}

	if err := s.downloadManager.Download(ctx, s.sourceUrl, s.ArchiveFilePath()); err != nil {
		log.FromContext(ctx).ErrorfWithFields(log.Fields{"err": err}, "failed to download remote archive to check updates for %s", s.Name)
		return s
	}

	hasUpdates, err := s.downloadManager.CheckUpdates(ctx, s.sourceUrl, s.ArchiveFilePath())
	if err != nil {
		log.FromContext(ctx).ErrorfWithFields(log.Fields{"err": err}, "failed to check for updates for %s", s.Name)
		return s
	}

	if !hasUpdates {
		return s
	}

	log.FromContext(ctx).Infof("Found updates for %s", s.Name)
	if err := s.downloadManager.Download(ctx, s.sourceUrl, s.ArchiveFilePath()); err != nil {
		log.FromContext(ctx).ErrorfWithFields(log.Fields{"err": err}, "failed to download updates for %s", s.Name)
		return s
	}
	if err := s.downloadManager.ApplyUpdate(ctx, s.ArchiveFilePath()); err != nil {
		log.FromContext(ctx).ErrorfWithFields(log.Fields{"err": err}, "failed to apply updates for %s", s.Name)
		return s
	}

	log.FromContext(ctx).Infof("Applied updates for %s", s.Name)

	return s
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

	if err := s.downloadManager.Download(ctx, s.sourceUrl, s.ArchiveFilePath()); err != nil {
		updates[s.Name] = &entity.UpdateStatus{Error: err.Error()}
		return updates, nil
	}

	hasUpdates, err := s.downloadManager.CheckUpdates(ctx, s.sourceUrl, s.ArchiveFilePath())
	if err != nil {
		updates[s.Name] = &entity.UpdateStatus{Error: err.Error()}
		return updates, nil
	}

	if hasUpdates {
		updates[s.Name] = &entity.UpdateStatus{Available: true}
		s.downloadManager.RemoveTmp(s.ArchiveFilePath())
	}

	return updates, nil
}

func (s *PatchesSource) DirPath() string {
	return s.dirPath
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

	if err := s.downloadManager.Download(ctx, s.sourceUrl, s.ArchiveFilePath()); err != nil {
		updates[s.Name] = &entity.UpdateStatus{Error: err.Error()}
		return updates, nil
	}

	hasUpdates, err := s.downloadManager.CheckUpdates(ctx, s.sourceUrl, s.ArchiveFilePath())
	if err != nil {
		updates[s.Name] = &entity.UpdateStatus{Error: err.Error()}
		return updates, nil
	}

	if !hasUpdates {
		return nil, nil
	}

	if err := s.downloadManager.ApplyUpdate(ctx, s.ArchiveFilePath()); err != nil {
		updates[s.Name] = &entity.UpdateStatus{Error: err.Error()}
		return updates, nil
	}

	log.FromContext(ctx).Infof("Applied updates for %s", s.Name)

	updates[s.Name] = &entity.UpdateStatus{}

	return updates, nil
}
