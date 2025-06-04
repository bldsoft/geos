package geonames

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/bldsoft/gost/log"
	"github.com/robfig/cron"
)

type StoragePatchesSource struct {
	c           *cron.Cron
	sourceUrl   string
	storagePath string
}

func NewStoragePatchesSource(sourceUrl, storagePath string, c *cron.Cron, autoUpdatePeriod string) *StoragePatchesSource {
	s := &StoragePatchesSource{
		sourceUrl:   sourceUrl,
		storagePath: storagePath,
		c:           c,
	}

	ctx := context.Background()

	if len(sourceUrl) == 0 {
		log.FromContext(ctx).Warn("Source for geonames storage patches is not set. You will not be able to check for %s patches updates and download them without providing source.")
		return s
	}

	exist, err := s.CheckUpdates(ctx)
	if err != nil {
		log.FromContext(ctx).ErrorWithFields(log.Fields{"err": err}, "Failed to check for updates for geonames storage patches")
	}

	if exist {
		log.FromContext(ctx).Infof("Updates are available for geonames storage patches")
		err = s.Download(ctx)
		if err != nil {
			log.FromContext(ctx).ErrorWithFields(log.Fields{"err": err}, "Failed to download patches for geonames storage")
		} else {
			log.FromContext(ctx).Infof("Successfully downloaded updates for geonames storage patches")
		}
	}

	if err := s.initAutoUpdates(ctx, autoUpdatePeriod); err != nil {
		log.FromContext(ctx).ErrorWithFields(log.Fields{"err": err}, "Failed to initialize auto updates for geonames storage patches")
	}
	return s
}

func (s *StoragePatchesSource) initAutoUpdates(ctx context.Context, period string) error {
	if period == "" || period == "0" {
		return nil
	}

	if s.sourceUrl == "" || s.storagePath == "" {
		return fmt.Errorf("missing required params")
	}

	return s.c.AddFunc(fmt.Sprintf("@every %sh", period), func() {
		log.FromContext(ctx).Info("Executing auto update for geonames storage patches")

		exist, err := s.CheckUpdates(ctx)
		if err != nil {
			log.FromContext(ctx).ErrorWithFields(log.Fields{"err": err}, "failed to check updates for geonames storage patches")
			return
		}

		if !exist {
			log.FromContext(ctx).Info("No updates found for geonames storage patches")
			return
		}

		log.FromContext(ctx).Info("Found updates for geonames storage patches")

		if err := s.Download(ctx); err != nil {
			log.FromContext(ctx).ErrorWithFields(log.Fields{"err": err}, "failed to download updates for geonames storage patches")
			return
		}

		log.FromContext(ctx).Info("Successfully downloaded updates for geonames storage patches")
	})
}

func (s *StoragePatchesSource) CheckUpdates(ctx context.Context) (bool, error) {
	if s.storagePath == "" {
		return false, fmt.Errorf("geonames storage path is not set, unable to check for patches updates")
	}

	if s.sourceUrl == "" {
		return false, fmt.Errorf("geonames storage source is not set, unable to check for patches updates")
	}

	return s.checkUpdates(ctx)
}

func (s *StoragePatchesSource) checkUpdates(ctx context.Context) (bool, error) {
	remoteContent, err := downloadAndParseTarGz(s.sourceUrl)
	if err != nil {
		return false, err
	}

	localContent := make(map[string]struct{})

	files, err := os.ReadDir(s.storagePath)
	if err != nil {
		return false, err
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if strings.HasPrefix(file.Name(), "geonames") && strings.HasSuffix(file.Name(), ".json") {
			localContent[file.Name()] = struct{}{}
		}
	}

	if len(remoteContent) != len(localContent) {
		return true, nil
	}

	for fileName, content := range remoteContent {
		if _, exists := localContent[fileName]; !exists {
			return true, nil
		}

		localFilePath := path.Join(s.storagePath, fileName)
		localFileContent, err := os.ReadFile(localFilePath)
		if err != nil {
			return false, err
		}

		if !bytes.Equal(localFileContent, content) {
			return true, nil
		}
	}

	return false, nil
}

func (s *StoragePatchesSource) Download(ctx context.Context, _ ...bool) error {
	if s.storagePath == "" {
		return fmt.Errorf("geonames storage path is not set, unable to check for patches updates")
	}

	if s.sourceUrl == "" {
		return fmt.Errorf("geonames storage source is not set, unable to check for patches updates")
	}

	log.FromContext(ctx).Info("Downloading geonames storage patches")

	return s.download(ctx)
}

func (s *StoragePatchesSource) download(ctx context.Context) error {
	remoteContent, err := downloadAndParseTarGz(s.sourceUrl)
	if err != nil {
		return fmt.Errorf("failed to download remote content: %w", err)
	}

	files, err := os.ReadDir(s.storagePath)
	if err != nil {
		return fmt.Errorf("failed to read local directory %s: %w", s.storagePath, err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if strings.HasPrefix(file.Name(), "geonames") && strings.HasSuffix(file.Name(), ".json") {
			if err := os.Remove(path.Join(s.storagePath, file.Name())); err != nil {
				return fmt.Errorf("failed to remove old file %s: %w", file.Name(), err)
			}
		}
	}

	for fileName, content := range remoteContent {
		if err := os.WriteFile(path.Join(s.storagePath, fileName), content, 0644); err != nil {
			return fmt.Errorf("failed to write new file %s: %w", fileName, err)
		}
	}

	return nil
}

func downloadAndParseTarGz(url string) (map[string][]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad response status: %s", resp.Status)
	}

	gzReader, err := gzip.NewReader(resp.Body)
	if err != nil {
		return nil, err
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)
	files := make(map[string][]byte)

	for {
		hdr, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if hdr.Typeflag != tar.TypeReg || !strings.HasSuffix(hdr.Name, ".json") {
			continue
		}

		var buf bytes.Buffer
		if _, err := io.Copy(&buf, tarReader); err != nil {
			return nil, err
		}

		files[path.Base(hdr.Name)] = buf.Bytes()
	}

	return files, nil
}
