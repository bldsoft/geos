package source

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/bldsoft/geos/pkg/entity"
	"github.com/bldsoft/gost/log"
)

type DownloadFunc func(ctx context.Context, url string, writer io.Writer) error
type UpdateCheckerFunc func(ctx context.Context, url string, targetPath string) (bool, error)
type ApplyUpdateFunc func(ctx context.Context, targetPath string) error

type DownloadManager struct {
	name              entity.Subject
	downloadFunc      DownloadFunc
	updateCheckerFunc UpdateCheckerFunc
	applyUpdateFunc   ApplyUpdateFunc
}

func NewDownloadManager(name entity.Subject) *DownloadManager {
	dm := &DownloadManager{
		name: name,
	}

	dm.downloadFunc = dm.defaultHttpDownload
	dm.updateCheckerFunc = dm.defaultUpdateChecker
	dm.applyUpdateFunc = dm.defaultApplyUpdate

	return dm
}

func NewCustomDownloadManager(name entity.Subject, downloadFunc DownloadFunc, updateCheckerFunc UpdateCheckerFunc, applyUpdateFunc ApplyUpdateFunc) *DownloadManager {
	dm := &DownloadManager{
		name: name,
	}

	if downloadFunc == nil {
		downloadFunc = dm.defaultHttpDownload
	}
	if updateCheckerFunc == nil {
		updateCheckerFunc = dm.defaultUpdateChecker
	}
	if applyUpdateFunc == nil {
		applyUpdateFunc = dm.defaultApplyUpdate
	}

	dm.downloadFunc = downloadFunc
	dm.updateCheckerFunc = updateCheckerFunc
	dm.applyUpdateFunc = applyUpdateFunc

	return dm
}

type DownloadResult struct {
	Downloaded bool
	Updated    bool
}

func (dm *DownloadManager) RecoverInterruptedDownloads(ctx context.Context, targetPath string, sourceUrl string) error {
	tmpPath := dm.tmpFilePath(targetPath)

	if _, err := os.Stat(tmpPath); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to check temporary file status: %w", err)
	}

	log.FromContext(ctx).Warnf("Found interrupted download for %s, attempting to recover", dm.name)

	if err := os.Remove(tmpPath); err != nil {
		log.FromContext(ctx).WarnWithFields(log.Fields{"err": err}, "Failed to remove temporary file, continuing anyway")
	}

	log.FromContext(ctx).Infof("Re-downloading %s to recover from interrupted download", dm.name)
	err := dm.Download(ctx, sourceUrl, targetPath)
	if err != nil {
		return fmt.Errorf("failed to re-download during recovery: %w", err)
	}

	return nil
}

func (dm *DownloadManager) Download(ctx context.Context, url string, targetPath string) error {
	tmpPath := dm.tmpFilePath(targetPath)

	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.Remove(tmpPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to clean up existing temporary file: %w", err)
	}

	tmpFile, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}

	downloadErr := dm.downloadFunc(ctx, url, tmpFile)
	tmpFile.Close()

	if downloadErr != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("download failed: %w", downloadErr)
	}

	return nil
}

func (dm *DownloadManager) CheckUpdates(ctx context.Context, url string, targetPath string) (bool, error) {
	return dm.updateCheckerFunc(ctx, url, targetPath)
}

func (dm *DownloadManager) defaultApplyUpdate(ctx context.Context, targetPath string) error {
	tmpPath := dm.tmpFilePath(targetPath)
	backupPath := dm.backupFilePath(targetPath)

	if _, err := os.Stat(tmpPath); err != nil {
		return fmt.Errorf("temporary file not found or not accessible: %w", err)
	}

	if _, err := os.Stat(targetPath); err == nil {
		if err := os.Rename(targetPath, backupPath); err != nil {
			return fmt.Errorf("failed to create backup of existing file: %w", err)
		}
	}

	if err := os.Rename(tmpPath, targetPath); err != nil {
		if _, backupExists := os.Stat(backupPath); backupExists == nil {
			if restoreErr := os.Rename(backupPath, targetPath); restoreErr != nil {
				return fmt.Errorf("failed to move tmp file and failed to restore backup: move_error=%w, restore_error=%v", err, restoreErr)
			}
		}
		return fmt.Errorf("failed to move tmp file to final location: %w", err)
	}

	if _, err := os.Stat(backupPath); err == nil {
		_ = os.Remove(backupPath)
	}

	return nil
}

func (dm *DownloadManager) ApplyUpdate(ctx context.Context, targetPath string) error {
	return dm.applyUpdateFunc(ctx, targetPath)
}

func (dm *DownloadManager) tmpFilePath(targetPath string) string {
	return targetPath + ".tmp"
}

func (dm *DownloadManager) backupFilePath(targetPath string) string {
	return targetPath + ".backup"
}

func (dm *DownloadManager) defaultHttpDownload(ctx context.Context, url string, writer io.Writer) error {
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
		return fmt.Errorf("bad response status: %s", resp.Status)
	}

	_, err = io.Copy(writer, resp.Body)
	return err
}

func (dm *DownloadManager) RemoveTmp(targetPath string) error {
	tmpPath := dm.tmpFilePath(targetPath)
	return os.Remove(tmpPath)
}

func (dm *DownloadManager) defaultUpdateChecker(ctx context.Context, url string, targetPath string) (hasUpdates bool, err error) {
	tmpPath := dm.tmpFilePath(targetPath)
	if hasUpdates, err = compareFiles(tmpPath, targetPath); err == nil && !hasUpdates {
		dm.RemoveTmp(targetPath)
	}
	return
}

func compareFiles(tmpPath, targetPath string) (bool, error) {
	tmpFile, err := os.Open(tmpPath)
	if err != nil {
		return false, err
	}
	defer tmpFile.Close()

	currentFile, err := os.Open(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			return true, nil
		}
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
