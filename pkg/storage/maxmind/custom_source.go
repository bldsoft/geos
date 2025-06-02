package maxmind

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

type DBPatchesSource struct {
	c         *cron.Cron
	sourceUrl string
	dbPath    string
	name      string
}

func (s *DBPatchesSource) CheckUpdates(ctx context.Context) (bool, error) {
	if s.dbPath == "" {
		return false, fmt.Errorf("%s database path is not set, unable to check for patches updates", s.name)
	}

	if s.sourceUrl == "" {
		return false, fmt.Errorf("%s patches source is not set, unable to check for updates", s.name)
	}

	return s.checkUpdates(ctx)
}

func (s *DBPatchesSource) Download(ctx context.Context, _ ...bool) error {
	if s.dbPath == "" {
		return fmt.Errorf("%s database path is not set, unable to check for patches updates", s.name)
	}

	if s.sourceUrl == "" {
		return fmt.Errorf("%s patches source is not set, unable to check for updates", s.name)
	}

	log.FromContext(ctx).Infof("Downloading %s patches", s.name)

	return s.download(ctx)
}

func NewCustomDBSource(sourceUrl, dbPath, name string, cron *cron.Cron, autoUpdatePeriod string) *DBPatchesSource {
	s := &DBPatchesSource{
		sourceUrl: sourceUrl,
		dbPath:    dbPath,
		c:         cron,
		name:      name,
	}

	ctx := context.Background()

	if len(sourceUrl) != 0 {
		exist, err := s.CheckUpdates(ctx)
		if err != nil {
			log.FromContext(ctx).Errorf("Failed to check for updates for %s patches: %v", s.name, err)
		}

		if exist {
			log.FromContext(ctx).Infof("Updates are available for %s patches", s.name)
			err = s.Download(ctx)
			if err != nil {
				log.FromContext(ctx).Errorf("Failed to download patches for %s: %v", s.name, err)
			} else {
				log.FromContext(ctx).Infof("Successfully downloaded updates for %s patches", s.name)
			}
		}

		if err := s.initAutoUpdates(ctx, autoUpdatePeriod); err != nil {
			log.FromContext(ctx).ErrorfWithFields(log.Fields{"err": err}, "Failed to initialize auto updates for %s patches", s.name)
		}
	} else {
		log.FromContext(ctx).Warnf("Source for %s patches is not set. You will NOT be able to check for %s patches updates and download them without providing source.", s.name, s.name)
	}

	return s
}

func (s *DBPatchesSource) initAutoUpdates(ctx context.Context, autoUpdatePeriod string) error {
	if autoUpdatePeriod == "" || autoUpdatePeriod == "0" {
		return nil
	}

	if s.sourceUrl == "" || s.dbPath == "" {
		return fmt.Errorf("missing required paths")
	}

	return s.c.AddFunc(fmt.Sprintf("@every %sh", autoUpdatePeriod), func() {
		log.FromContext(ctx).Infof("Executing auto update for %s patches", s.name)

		exist, err := s.CheckUpdates(ctx)
		if err != nil {
			log.FromContext(ctx).ErrorfWithFields(log.Fields{"err": err}, "failed to check for updates for %s patches", s.name)
			return
		}

		if !exist {
			log.FromContext(ctx).Infof("No updates found for %s patches", s.name)
			return
		}

		log.FromContext(ctx).Infof("Found update for %s patches", s.name)

		if err := s.Download(ctx); err != nil {
			log.FromContext(ctx).ErrorfWithFields(log.Fields{"err": err}, "failed to download updates for %s patches", s.name)
			return
		}

		log.FromContext(ctx).Infof("Successfully downloaded updates for %s patches", s.name)
	})
}

func (s *DBPatchesSource) download(ctx context.Context) error {
	remoteContent, err := downloadAndParseTarGz(s.sourceUrl)
	if err != nil {
		return fmt.Errorf("failed to download remote content: %w", err)
	}

	files, err := os.ReadDir(s.dbPath)
	if err != nil {
		return fmt.Errorf("failed to read local directory %s: %w", s.dbPath, err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if strings.HasPrefix(file.Name(), s.name) && strings.HasSuffix(file.Name(), ".json") {
			if err := os.Remove(path.Join(s.dbPath, file.Name())); err != nil {
				return fmt.Errorf("failed to remove old file %s: %w", file.Name(), err)
			}
		}
	}

	for fileName, content := range remoteContent {
		if err := os.WriteFile(path.Join(s.dbPath, fileName), content, 0644); err != nil {
			return fmt.Errorf("failed to write new file %s: %w", fileName, err)
		}
	}

	return nil
}

func (s *DBPatchesSource) checkUpdates(ctx context.Context) (bool, error) {
	remoteContent, err := downloadAndParseTarGz(s.sourceUrl)
	if err != nil {
		return false, err
	}

	localContent := make(map[string]struct{})

	files, err := os.ReadDir(s.dbPath)
	if err != nil {
		return false, err
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if strings.HasPrefix(file.Name(), s.name) && strings.HasSuffix(file.Name(), ".json") {
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

		localFilePath := path.Join(s.dbPath, fileName)
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
