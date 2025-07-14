package source

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/bldsoft/geos/pkg/entity"
	"github.com/bldsoft/geos/pkg/storage/maxmind/mmdb"
	"github.com/bldsoft/gost/log"
	"github.com/hashicorp/go-version"
)

const metadataChunkSize = 128 * 1024

type MaxmindSource struct {
	sourceUrl       string
	dbPath          string
	name            entity.Subject
	downloadManager *DownloadManager
}

type mmdbState struct {
	version    *version.Version
	buildEpoch uint
}

func (s *mmdbState) isNewerThan(other *mmdbState) bool {
	if s.version.GreaterThan(other.version) {
		return true
	}

	return s.buildEpoch > other.buildEpoch
}

func (s *MaxmindSource) checkUpdates(ctx context.Context, rep url string, _ string) (bool, error) {
	res, err := s.downloadRange(ctx, url, -metadataChunkSize)
	if err != nil {
		return false, err
	}

	remoteState, err := s.extractMmdbState(res)
	if err != nil {
		return false, fmt.Errorf("failed to extract version from response: %w", err)
	}

	localState, err := s.fileMetadataMmdbState(s.dbPath)
	if err != nil {
		return false, fmt.Errorf("failed to extract version from local file: %w", err)
	}

	return remoteState.isNewerThan(localState), nil
}

func (s *MaxmindSource) download(ctx context.Context, url string, writer io.Writer) error {
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

	_, err = io.Copy(writer, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func NewMMDBSource(sourceUrl, dbPath string, name entity.Subject, autoUpdatePeriod int) *MaxmindSource {
	s := &MaxmindSource{
		sourceUrl: sourceUrl,
		dbPath:    dbPath,
		name:      name,
	}

	s.downloadManager = NewCustomDownloadManager(name, s.download, s.checkUpdates, nil)

	ctx := context.Background()

	if len(sourceUrl) == 0 {
		log.FromContext(ctx).Warnf("Source for %s database is not set, assuming it is already downloaded. You will NOT be able to check for database updates and download them without providing source.", name)
		return s
	}

	if err := s.downloadManager.RecoverInterruptedDownloads(ctx, dbPath, sourceUrl); err != nil {
		log.FromContext(ctx).ErrorfWithFields(log.Fields{"err": err}, "failed to handle interrupted downloads for %s database", name)
	}

	f, err := os.Stat(dbPath)
	if err != nil {
		if !os.IsNotExist(err) {
			log.FromContext(ctx).ErrorfWithFields(log.Fields{"err": err}, "failed to access %s database at %s", name, dbPath)
			return s
		}

		log.FromContext(ctx).Warnf("No %s database found at %s, downloading from source", name, dbPath)

		if err := s.downloadManager.Download(ctx, sourceUrl, dbPath); err != nil {
			log.FromContext(ctx).ErrorfWithFields(log.Fields{"err": err}, "failed to download %s database from provided source", name)
			return s
		}

		if err := s.downloadManager.ApplyUpdate(ctx, dbPath); err != nil {
			log.FromContext(ctx).ErrorfWithFields(log.Fields{"err": err}, "failed to update local files for %s database", name)
			return s
		}
	}

	if f == nil || autoUpdatePeriod == 0 {
		return s
	}

	if err := s.downloadManager.Download(ctx, sourceUrl, s.dbPath); err != nil {
		log.FromContext(ctx).ErrorfWithFields(log.Fields{"err": err}, "failed to download %s database from provided source", name)
		return s
	}

	exist, err := s.downloadManager.CheckUpdates(ctx, sourceUrl, s.dbPath)
	if err != nil {
		log.FromContext(ctx).ErrorfWithFields(log.Fields{"err": err}, "failed to check for updates for %s database", name)
		return s
	}

	if !exist {
		s.downloadManager.RemoveTmp(s.dbPath)
		return s
	}

	log.FromContext(ctx).Infof("Found updates for %s database", name)

	if err := s.downloadManager.ApplyUpdate(ctx, s.dbPath); err != nil {
		log.FromContext(ctx).ErrorfWithFields(log.Fields{"err": err}, "failed to update %s database files: %v", name, err)
	}

	log.FromContext(ctx).Infof("Applied updates for %s database", name)

	return s
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

	exist, err := s.downloadManager.CheckUpdates(ctx, s.sourceUrl, s.dbPath)
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

	exist, err := s.downloadManager.CheckUpdates(ctx, s.sourceUrl, s.dbPath)
	if err != nil {
		updates[s.name] = &entity.UpdateStatus{Error: err.Error()}
		return updates, nil
	}

	if !exist {
		return nil, nil
	}

	if err := s.downloadManager.Download(ctx, s.sourceUrl, s.dbPath); err != nil {
		updates[s.name] = &entity.UpdateStatus{Error: err.Error()}
		return updates, nil
	}

	if err := s.downloadManager.ApplyUpdate(ctx, s.dbPath); err != nil {
		updates[s.name] = &entity.UpdateStatus{Error: err.Error()}
	}

	updates[s.name] = &entity.UpdateStatus{}

	return updates, nil
}

func (s *MaxmindSource) extractMmdbState(metadataBuf []byte) (*mmdbState, error) {
	meta, err := mmdb.DecodeMetadata(metadataBuf)
	if err != nil {
		return nil, fmt.Errorf("metadata decoding failed: %w", err)
	}

	v, err := version.NewVersion(fmt.Sprintf("%d.%d", meta.BinaryFormatMajorVersion, meta.BinaryFormatMinorVersion))
	if err != nil {
		return nil, err
	}

	return &mmdbState{
		version:    v,
		buildEpoch: meta.BuildEpoch,
	}, nil
}

func (s *MaxmindSource) fileMetadataMmdbState(path string) (*mmdbState, error) {
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

	return s.extractMmdbState(buffer)
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
		return nil, fmt.Errorf("server did not return partial content: %s", resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}
