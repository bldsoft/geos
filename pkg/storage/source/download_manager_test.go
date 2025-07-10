package source

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/bldsoft/geos/pkg/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDownloadManager_Download_Success(t *testing.T) {
	testContent := "test file content"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(testContent))
	}))
	defer server.Close()

	tempDir := t.TempDir()
	targetPath := filepath.Join(tempDir, "test.txt")

	dm := NewDownloadManager(entity.Subject("test"))

	err := dm.Download(context.Background(), server.URL, targetPath)
	require.NoError(t, err)

	tmpPath := dm.tmpFilePath(targetPath)
	content, err := os.ReadFile(tmpPath)
	require.NoError(t, err)
	assert.Equal(t, testContent, string(content))

	_, err = os.Stat(targetPath)
	assert.True(t, os.IsNotExist(err))
}

func TestDownloadManager_Download_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	tempDir := t.TempDir()
	targetPath := filepath.Join(tempDir, "test.txt")

	dm := NewDownloadManager(entity.Subject("test"))

	err := dm.Download(context.Background(), server.URL, targetPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "bad response status")

	tmpPath := dm.tmpFilePath(targetPath)
	_, err = os.Stat(tmpPath)
	assert.True(t, os.IsNotExist(err))
}

func TestDownloadManager_ApplyUpdate_Success(t *testing.T) {
	tempDir := t.TempDir()
	targetPath := filepath.Join(tempDir, "test.txt")

	existingContent := "existing content"
	err := os.WriteFile(targetPath, []byte(existingContent), 0644)
	require.NoError(t, err)

	dm := NewDownloadManager(entity.Subject("test"))

	tmpPath := dm.tmpFilePath(targetPath)
	newContent := "new content"
	err = os.WriteFile(tmpPath, []byte(newContent), 0644)
	require.NoError(t, err)

	err = dm.ApplyUpdate(context.Background(), targetPath)
	require.NoError(t, err)

	content, err := os.ReadFile(targetPath)
	require.NoError(t, err)
	assert.Equal(t, newContent, string(content))

	_, err = os.Stat(tmpPath)
	assert.True(t, os.IsNotExist(err))

	backupPath := dm.backupFilePath(targetPath)
	_, err = os.Stat(backupPath)
	assert.True(t, os.IsNotExist(err))
}

func TestDownloadManager_CheckForUpdates_HasUpdate(t *testing.T) {
	newContent := "new content"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(newContent))
	}))
	defer server.Close()

	tempDir := t.TempDir()
	targetPath := filepath.Join(tempDir, "test.txt")

	err := os.WriteFile(targetPath, []byte("old content"), 0644)
	require.NoError(t, err)

	dm := NewDownloadManager(entity.Subject("test"))

	// First download to create the temporary file
	err = dm.Download(context.Background(), server.URL, targetPath)
	require.NoError(t, err)

	hasUpdate, err := dm.CheckUpdates(context.Background(), server.URL, targetPath)
	require.NoError(t, err)
	assert.True(t, hasUpdate)
}

func TestDownloadManager_CheckForUpdates_NoUpdate(t *testing.T) {
	content := "same content"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(content))
	}))
	defer server.Close()

	tempDir := t.TempDir()
	targetPath := filepath.Join(tempDir, "test.txt")

	err := os.WriteFile(targetPath, []byte(content), 0644)
	require.NoError(t, err)

	dm := NewDownloadManager(entity.Subject("test"))

	// First download to create the temporary file
	err = dm.Download(context.Background(), server.URL, targetPath)
	require.NoError(t, err)

	hasUpdate, err := dm.CheckUpdates(context.Background(), server.URL, targetPath)
	require.NoError(t, err)
	assert.False(t, hasUpdate)
}

func TestDownloadManager_RecoverInterruptedDownloads_WithTempFile(t *testing.T) {
	testContent := "recovered content"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(testContent))
	}))
	defer server.Close()

	tempDir := t.TempDir()
	targetPath := filepath.Join(tempDir, "test.txt")

	dm := NewDownloadManager(entity.Subject("test"))

	tmpPath := dm.tmpFilePath(targetPath)
	err := os.MkdirAll(filepath.Dir(tmpPath), 0755)
	require.NoError(t, err)
	err = os.WriteFile(tmpPath, []byte("incomplete"), 0644)
	require.NoError(t, err)

	err = dm.RecoverInterruptedDownloads(context.Background(), targetPath, server.URL)
	require.NoError(t, err)

	content, err := os.ReadFile(targetPath)
	require.NoError(t, err)
	assert.Equal(t, testContent, string(content))

	_, err = os.Stat(tmpPath)
	assert.True(t, os.IsNotExist(err))
}

func TestDownloadManager_RecoverInterruptedDownloads_NoTempFile(t *testing.T) {
	tempDir := t.TempDir()
	targetPath := filepath.Join(tempDir, "test.txt")

	dm := NewDownloadManager(entity.Subject("test"))

	err := dm.RecoverInterruptedDownloads(context.Background(), targetPath, "http://example.com")
	require.NoError(t, err)
}

func TestDownloadManager_CustomFunctions(t *testing.T) {
	downloadCalled := false
	updateCheckCalled := false

	customDownload := func(ctx context.Context, url string, writer io.Writer) error {
		downloadCalled = true
		_, err := writer.Write([]byte("custom content"))
		return err
	}

	customUpdateChecker := func(ctx context.Context, url string, targetPath string) (bool, error) {
		updateCheckCalled = true
		return true, nil
	}

	dm := NewCustomDownloadManager(entity.Subject("custom"), customDownload, customUpdateChecker, nil)

	tempDir := t.TempDir()
	targetPath := filepath.Join(tempDir, "test.txt")

	err := dm.Download(context.Background(), "http://example.com", targetPath)
	require.NoError(t, err)
	assert.True(t, downloadCalled)

	hasUpdate, err := dm.CheckUpdates(context.Background(), "http://example.com", targetPath)
	require.NoError(t, err)
	assert.True(t, hasUpdate)
	assert.True(t, updateCheckCalled)
}
