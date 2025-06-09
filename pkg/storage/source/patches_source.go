package source

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

	"github.com/bldsoft/geos/pkg/entity"
	"github.com/bldsoft/gost/log"
	"github.com/robfig/cron"
)

type PatchesSource struct {
	c         *cron.Cron
	sourceUrl string
	dirPath   string
	prefix    string
	Name      entity.Subject
}

func NewPatchesSource(sourceUrl, dirPath, prefix string, name entity.Subject, cron *cron.Cron, autoUpdatePeriod string) *PatchesSource {
	s := &PatchesSource{
		sourceUrl: sourceUrl,
		dirPath:   dirPath,
		c:         cron,
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

	remoteContent, err := getFiles(s.sourceUrl)
	if err != nil {
		log.FromContext(ctx).Errorf("Failed to get source files for %s: %v", s.Name, err)
		return s
	}

	exist, err := s.checkUpdates(remoteContent)
	if err != nil {
		log.FromContext(ctx).Errorf("Failed to check for updates for %s: %v", s.Name, err)
		return s
	}

	if exist {
		log.FromContext(ctx).Infof("Updates are available for %s", s.Name)
		err = s.updateLocalFiles(remoteContent)
		if err != nil {
			log.FromContext(ctx).Errorf("Failed to apply updates for %s: %v", s.Name, err)
		} else {
			log.FromContext(ctx).Infof("Successfully applied updates for %s", s.Name)
		}
	}

	return s
}

func (s *PatchesSource) initAutoUpdates(ctx context.Context, autoUpdatePeriod string) error {
	if autoUpdatePeriod == "" || autoUpdatePeriod == "0" {
		return nil
	}

	if s.sourceUrl == "" || s.dirPath == "" {
		return fmt.Errorf("missing required params")
	}

	return s.c.AddFunc(fmt.Sprintf("@every %sh", autoUpdatePeriod), func() {
		log.FromContext(ctx).Infof("Executing auto update for %s", s.Name)

		remoteContent, err := getFiles(s.sourceUrl)
		if err != nil {
			log.FromContext(ctx).ErrorfWithFields(log.Fields{"err": err}, "failed to get source files for %s", s.Name)
			return
		}

		exist, err := s.checkUpdates(remoteContent)
		if err != nil {
			log.FromContext(ctx).ErrorfWithFields(log.Fields{"err": err}, "failed to check for updates for %s", s.Name)
			return
		}

		if !exist {
			log.FromContext(ctx).Infof("No updates found for %s", s.Name)
			return
		}

		log.FromContext(ctx).Infof("Found update for %s", s.Name)

		if err := s.updateLocalFiles(remoteContent); err != nil {
			log.FromContext(ctx).ErrorfWithFields(log.Fields{"err": err}, "failed to download updates for %s", s.Name)
			return
		}

		log.FromContext(ctx).Infof("Successfully applied updates for %s", s.Name)
	})
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

	remoteContent, err := getFiles(s.sourceUrl)
	if err != nil {
		updates[s.Name] = &entity.UpdateStatus{Error: err.Error()}
		return updates, nil
	}

	exist, err := s.checkUpdates(remoteContent)
	if err != nil {
		updates[s.Name] = &entity.UpdateStatus{Error: err.Error()}
		return updates, nil
	}

	if exist {
		updates[s.Name] = &entity.UpdateStatus{Available: true}
	}

	return updates, nil
}

func (s *PatchesSource) checkUpdates(remoteContent map[string][]byte) (bool, error) {
	localContent := make(map[string]struct{})

	files, err := os.ReadDir(s.dirPath)
	if err != nil {
		return false, err
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if strings.HasPrefix(file.Name(), s.prefix) && strings.HasSuffix(file.Name(), ".json") {
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

		localFilePath := path.Join(s.dirPath, fileName)
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

func (s *PatchesSource) Download(ctx context.Context, _ ...bool) (entity.Updates, error) {
	updates := entity.Updates{}

	if s.dirPath == "" {
		updates[s.Name] = &entity.UpdateStatus{Error: fmt.Errorf("%s path is not set, unable to check for updates", s.Name).Error()}
		return updates, nil
	}

	if s.sourceUrl == "" {
		updates[s.Name] = &entity.UpdateStatus{Error: fmt.Errorf("%s source is not set, unable to check for updates", s.Name).Error()}
		return updates, nil
	}

	remoteContent, err := getFiles(s.sourceUrl)
	if err != nil {
		updates[s.Name] = &entity.UpdateStatus{Error: err.Error()}
		return updates, nil
	}

	exist, err := s.checkUpdates(remoteContent)
	if err != nil {
		updates[s.Name] = &entity.UpdateStatus{Error: err.Error()}
		return updates, nil
	}

	if !exist {
		return nil, nil
	}

	if err := s.updateLocalFiles(remoteContent); err != nil {
		updates[s.Name] = &entity.UpdateStatus{Error: err.Error()}
		return updates, nil
	}

	log.FromContext(ctx).Infof("Applied updates for %s", s.Name)

	updates[s.Name] = &entity.UpdateStatus{}

	return updates, nil
}

func (s *PatchesSource) updateLocalFiles(remoteContent map[string][]byte) error {
	files, err := os.ReadDir(s.dirPath)
	if err != nil {
		return fmt.Errorf("failed to read local directory %s: %w", s.dirPath, err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if strings.HasPrefix(file.Name(), s.prefix) && strings.HasSuffix(file.Name(), ".json") {
			if err := os.Remove(path.Join(s.dirPath, file.Name())); err != nil {
				return fmt.Errorf("failed to remove old file %s: %w", file.Name(), err)
			}
		}
	}

	for fileName, content := range remoteContent {
		if err := os.WriteFile(path.Join(s.dirPath, fileName), content, 0644); err != nil {
			return fmt.Errorf("failed to write new file %s: %w", fileName, err)
		}
	}

	return nil
}

func (s *PatchesSource) DirPath() string {
	return s.dirPath
}

func getFiles(url string) (map[string][]byte, error) {
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
