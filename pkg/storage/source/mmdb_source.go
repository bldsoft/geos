package source

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/bldsoft/geos/pkg/entity"
	"github.com/bldsoft/geos/pkg/storage/maxmind/mmdb"
	"github.com/bldsoft/gost/log"
	"github.com/hashicorp/go-version"
)

const metadataChunkSize = 128 * 1024

type MaxmindSource struct {
	sourceUrl string
	dbPath    string
	name      entity.Subject
}

func (s *MaxmindSource) tmpFilePath() string {
	return fmt.Sprintf("%s.tmp", s.dbPath)
}

func (s *MaxmindSource) CheckUpdates(ctx context.Context) (entity.Updates, error) {
	updates := entity.Updates{}

	if s.dbPath == "" {
		updates[s.name] = &entity.UpdateStatus{Error: fmt.Errorf("path for %s database is not set", s.name).Error()}
		return updates, nil
	}

	if s.sourceUrl == "" {
		updates[s.name] = &entity.UpdateStatus{Error: fmt.Errorf("source for %s database is not set", s.name).Error()}
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

func (s *MaxmindSource) Download(ctx context.Context) (entity.Updates, error) {
	updates := entity.Updates{}

	if s.dbPath == "" {
		updates[s.name] = &entity.UpdateStatus{Error: fmt.Errorf("path for %s database is not set", s.name).Error()}
		return updates, nil
	}

	if s.sourceUrl == "" {
		updates[s.name] = &entity.UpdateStatus{Error: fmt.Errorf("source for %s database is not set", s.name).Error()}
		return updates, nil
	}

	exist, err := s.checkUpdates(ctx)
	if err != nil {
		updates[s.name] = &entity.UpdateStatus{Error: err.Error()}
		return updates, err
	}

	if !exist {
		return nil, nil
	}

	if err := s.downloadTmp(ctx); err != nil {
		updates[s.name] = &entity.UpdateStatus{Error: err.Error()}
	}

	if err := s.updateLocalFiles(); err != nil {
		updates[s.name] = &entity.UpdateStatus{Error: err.Error()}
	}

	return updates, nil
}

func (s *MaxmindSource) updateLocalFiles() error {
	if _, err := os.Stat(s.tmpFilePath()); os.IsNotExist(err) {
		return err
	}

	if err := os.Rename(s.tmpFilePath(), s.dbPath); err != nil {
		return fmt.Errorf("failed to rename temporary file to %s: %w", s.dbPath, err)
	}

	return os.Remove(s.tmpFilePath())
}

func NewMMDBSource(sourceUrl, dbPath string, name entity.Subject, autoUpdatePeriod int) *MaxmindSource {
	s := &MaxmindSource{
		sourceUrl: sourceUrl,
		dbPath:    dbPath,
		name:      name,
	}

	ctx := context.Background()

	if len(sourceUrl) == 0 {
		log.FromContext(ctx).Warnf("Source for %s database is not set, assuming it is already downloaded. You will NOT be able to check for database updates and download them without providing source.", name)
		return s
	}

	if err := s.initAutoUpdates(ctx, autoUpdatePeriod); err != nil {
		log.FromContext(ctx).ErrorfWithFields(log.Fields{"err": err}, "failed to initialize auto updates for %s database", name)
	}

	f, err := os.Stat(dbPath)
	if err != nil {
		if !os.IsNotExist(err) {
			log.FromContext(ctx).ErrorfWithFields(log.Fields{"err": err}, "failed to access %s database at %s", name, dbPath)
			return s
		}

		log.FromContext(ctx).Warnf("No %s database found at %s, downloading from source", name, dbPath)

		if err := s.downloadTmp(ctx); err != nil {
			log.FromContext(ctx).ErrorfWithFields(log.Fields{"err": err}, "failed to download %s database from provided source", name)
			return s
		}

		if err := s.updateLocalFiles(); err != nil {
			log.FromContext(ctx).ErrorfWithFields(log.Fields{"err": err}, "failed to update local files for %s database", name)
			return s
		}
	}

	if f == nil || autoUpdatePeriod == 0 {
		return s
	}

	exist, err := s.checkUpdates(ctx)
	if err != nil {
		log.FromContext(ctx).ErrorfWithFields(log.Fields{"err": err}, "failed to check for updates for %s database", name)
		return s
	}

	if !exist {
		return s
	}

	if err := s.downloadTmp(ctx); err != nil {
		log.FromContext(ctx).ErrorfWithFields(log.Fields{"err": err}, "failed to download updates for %s database", name)
		return s
	}

	if err := s.updateLocalFiles(); err != nil {
		log.FromContext(ctx).ErrorfWithFields(log.Fields{"err": err}, "failed to update %s database files: %v", name, err)
	}

	return s
}

func (s *MaxmindSource) initAutoUpdates(ctx context.Context, autoUpdatePeriod int) error {
	if autoUpdatePeriod == 0 {
		return nil
	}

	if s.sourceUrl == "" || s.dbPath == "" {
		return fmt.Errorf("missing required paths")
	}

	go func() {
		timer := time.NewTicker(time.Duration(autoUpdatePeriod) * time.Hour)
		defer timer.Stop()

		for range timer.C {
			log.FromContext(ctx).Infof("Executing auto update for %s", s.name)

			available, err := s.checkUpdates(ctx)
			if err != nil {
				log.FromContext(ctx).ErrorfWithFields(log.Fields{"err": err}, "failed to run auto update for %s", s.name)
				return
			}

			if !available {
				log.FromContext(ctx).Infof("No updates found during automatic check for %s", s.name)
				return
			}

			log.FromContext(ctx).Infof("Found updates during automatic check for %s", s.name)

			if err := s.downloadTmp(ctx); err != nil {
				log.FromContext(ctx).ErrorfWithFields(log.Fields{"err": err}, "failed to download updates for %s", s.name)
				return
			}

			log.FromContext(ctx).Infof("Updates applied for %s", s.name)
		}
	}()

	return nil
}

func (s *MaxmindSource) DirPath() string {
	return s.dbPath
}

func (s *MaxmindSource) extractVersion(metadataBuf []byte) (*version.Version, error) {
	meta, err := mmdb.DecodeMetadata(metadataBuf)
	if err != nil {
		return nil, fmt.Errorf("metadata decoding failed: %w", err)
	}

	return version.NewVersion(fmt.Sprintf("%d.%d", meta.BinaryFormatMajorVersion, meta.BinaryFormatMinorVersion))
}

func (s *MaxmindSource) fileMetadataVersion(path string) (*version.Version, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}
	fileSize := fileInfo.Size()

	startOffset := fileSize - metadataChunkSize
	if fileSize < metadataChunkSize {
		startOffset = 0
	}

	readSize := fileSize - startOffset
	buffer := make([]byte, readSize)
	_, err = file.ReadAt(buffer, startOffset)
	if err != nil {
		return nil, fmt.Errorf("failed to read file chunk: %w", err)
	}

	return s.extractVersion(buffer)
}

func (s *MaxmindSource) downloadRange(ctx context.Context, sourceUrl string, chunkSize int) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", sourceUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Range", fmt.Sprintf("bytes=%d", chunkSize))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusPartialContent {
		return nil, fmt.Errorf("server does not support range requests")
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (s *MaxmindSource) checkUpdates(ctx context.Context) (bool, error) {
	res, err := s.downloadRange(ctx, s.sourceUrl, -metadataChunkSize)
	if err != nil {
		return false, err
	}

	remoteVersion, err := s.extractVersion(res)
	if err != nil {
		return false, fmt.Errorf("failed to extract version from response: %w", err)
	}

	localVersion, err := s.fileMetadataVersion(s.dbPath)
	if err != nil {
		return false, fmt.Errorf("failed to extract version from local file: %w", err)
	}

	return remoteVersion.GreaterThan(localVersion), nil
}

func (s *MaxmindSource) downloadTmp(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", s.sourceUrl, nil)
	if err != nil {
		return err
	}

	tmpFile, err := os.Create(s.tmpFilePath())
	if err != nil {
		return err
	}
	defer tmpFile.Close()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download file: %s", resp.Status)
	}

	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
