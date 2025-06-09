package source

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"

	"github.com/bldsoft/geos/pkg/entity"
	"github.com/bldsoft/geos/pkg/storage/maxmind/mmdb"
	"github.com/bldsoft/gost/log"
	"github.com/hashicorp/go-version"
	"github.com/oschwald/maxminddb-golang"
	"github.com/robfig/cron"
)

const metadataChunkSize = 128 * 1024

var metadataStartMarker = []byte("\xAB\xCD\xEFMaxMind.com")

type MaxmindSource struct {
	c         *cron.Cron
	sourceUrl string
	dbPath    string
	name      entity.Subject
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

func (s *MaxmindSource) Download(ctx context.Context, update ...bool) (entity.Updates, error) {
	updates := entity.Updates{}

	if s.dbPath == "" {
		updates[s.name] = &entity.UpdateStatus{Error: fmt.Errorf("path for %s database is not set", s.name).Error()}
		return updates, nil
	}

	if s.sourceUrl == "" {
		updates[s.name] = &entity.UpdateStatus{Error: fmt.Errorf("source for %s database is not set", s.name).Error()}
		return updates, nil
	}

	if len(update) == 0 || !update[0] {
		if _, err := os.Stat(s.dbPath); err == nil {
			log.FromContext(ctx).Infof("%s database already exists, skipping download", s.name)
			return nil, nil
		}
	}

	exist, err := s.checkUpdates(ctx)
	if err != nil {
		updates[s.name] = &entity.UpdateStatus{Error: err.Error()}
		return updates, err
	}

	if !exist {
		return nil, nil
	}

	if err := s.download(ctx); err != nil {
		updates[s.name] = &entity.UpdateStatus{Error: err.Error()}
	} else {
		updates[s.name] = &entity.UpdateStatus{}
	}

	return updates, nil
}

func NewMMDBSource(sourceUrl, dbPath string, name entity.Subject, cron *cron.Cron, autoUpdatePeriod string) *MaxmindSource {
	s := &MaxmindSource{
		sourceUrl: sourceUrl,
		dbPath:    dbPath,
		name:      name,
		c:         cron,
	}

	ctx := context.Background()

	if len(sourceUrl) != 0 {
		if err := s.download(ctx); err != nil {
			log.FromContext(ctx).ErrorfWithFields(log.Fields{"err": err}, "failed to download %s database from provided source", name)
		}

		if err := s.initAutoUpdates(ctx, autoUpdatePeriod); err != nil {
			log.FromContext(ctx).ErrorfWithFields(log.Fields{"err": err}, "failed to initialize auto updates for %s database", name)
		}
	} else {
		log.FromContext(ctx).Warnf("Source for %s database is not set, assuming it is already downloaded. You will NOT be able to check for database updates and download them without providing source.", name)
	}

	return s
}

func (s *MaxmindSource) initAutoUpdates(ctx context.Context, autoUpdatePeriod string) error {
	if autoUpdatePeriod == "" || autoUpdatePeriod == "0" {
		return nil
	}

	if s.sourceUrl == "" || s.dbPath == "" {
		return fmt.Errorf("missing required paths")
	}

	return s.c.AddFunc(fmt.Sprintf("@every %sh", autoUpdatePeriod), func() {
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

		if err := s.download(ctx); err != nil {
			log.FromContext(ctx).ErrorfWithFields(log.Fields{"err": err}, "failed to download updates for %s", s.name)
			return
		}

		log.FromContext(ctx).Infof("Updates applied for %s", s.name)
	})
}

func (s *MaxmindSource) DirPath() string {
	return s.dbPath
}

func (s *MaxmindSource) extractVersion(metadataBuf []byte) (*version.Version, error) {
	metadataDecoder := mmdb.Decoder{Buffer: metadataBuf}

	var metadata maxminddb.Metadata
	rvMetadata := reflect.ValueOf(&metadata)
	_, err := metadataDecoder.Decode(0, rvMetadata, 0)
	if err != nil {
		return nil, fmt.Errorf("metadata decoding failed: %w", err)
	}

	return version.NewVersion(fmt.Sprintf("%d.%d", metadata.BinaryFormatMajorVersion, metadata.BinaryFormatMinorVersion))
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

	metadataIndex := bytes.LastIndex(buffer, metadataStartMarker)
	if metadataIndex == -1 {
		return nil, fmt.Errorf("metadata marker not found")
	}

	metadataStart := metadataIndex + len(metadataStartMarker)

	return s.extractVersion(buffer[metadataStart:])
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

	metadataStart := bytes.LastIndex(res, metadataStartMarker)

	if metadataStart == -1 {
		return false, fmt.Errorf("metadata start marker not found in the response")
	}

	metadataStart += len(metadataStartMarker)

	remoteVersion, err := s.extractVersion(res[metadataStart:])
	if err != nil {
		return false, fmt.Errorf("failed to extract version from response: %w", err)
	}

	localVersion, err := s.fileMetadataVersion(s.dbPath)
	if err != nil {
		return false, fmt.Errorf("failed to extract version from local file: %w", err)
	}

	return remoteVersion.GreaterThan(localVersion), nil
}

func (s *MaxmindSource) download(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", s.sourceUrl, nil)
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

	file, err := os.Create(s.dbPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
