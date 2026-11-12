package source

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOpenPatchFile(t *testing.T) {
	t.Parallel()

	remote, err := OpenPatchFile("https://example.com/patches/city_patch.tar.gz", "/var/lib/geos", "city")
	require.NoError(t, err)
	require.Equal(t, "/var/lib/geos/city_patch.gz", remote.LocalPath)
	require.Equal(t, "https://example.com/patches/city_patch.tar.gz", remote.RemoteURL)

	upper, err := OpenPatchFile("HTTPS://example.com/patches/city_patch.tar.gz", "/var/lib/geos", "city")
	require.NoError(t, err)
	require.Equal(t, "/var/lib/geos/city_patch.gz", upper.LocalPath)
	require.Equal(t, "HTTPS://example.com/patches/city_patch.tar.gz", upper.RemoteURL)

	dir := t.TempDir()
	localPath := filepath.Join(dir, "city_patch.gz")
	require.NoError(t, os.WriteFile(localPath, []byte("patch"), 0o644))

	local, err := OpenPatchFile(localPath, "/var/lib/geos", "city")
	require.NoError(t, err)
	require.Equal(t, localPath, local.LocalPath)
	require.Empty(t, local.RemoteURL)

	fromFileURL, err := OpenPatchFile("file://"+localPath, "/var/lib/geos", "city")
	require.NoError(t, err)
	require.Equal(t, localPath, fromFileURL.LocalPath)
	require.Empty(t, fromFileURL.RemoteURL)

	fromLocalhost, err := OpenPatchFile("file://localhost"+localPath, "/var/lib/geos", "city")
	require.NoError(t, err)
	require.Equal(t, localPath, fromLocalhost.LocalPath)
	require.Empty(t, fromLocalhost.RemoteURL)

	percentDir := filepath.Join(dir, "50%")
	require.NoError(t, os.Mkdir(percentDir, 0o755))
	percentPath := filepath.Join(percentDir, "city_patch.json")
	require.NoError(t, os.WriteFile(percentPath, []byte("patch"), 0o644))
	percent, err := OpenPatchFile(percentPath, "/var/lib/geos", "city")
	require.NoError(t, err)
	require.Equal(t, percentPath, percent.LocalPath)
	require.Empty(t, percent.RemoteURL)

	colonDir := filepath.Join(dir, "patches:v2")
	require.NoError(t, os.Mkdir(colonDir, 0o755))
	colonPath := filepath.Join(colonDir, "city.json")
	require.NoError(t, os.WriteFile(colonPath, []byte("patch"), 0o644))
	colon, err := OpenPatchFile(colonPath, "/var/lib/geos", "city")
	require.NoError(t, err)
	require.Equal(t, colonPath, colon.LocalPath)

	_, err = OpenPatchFile("file://patches/city.json", "/var/lib/geos", "city")
	require.ErrorContains(t, err, "unsupported host")

	_, err = OpenPatchFile("file://", "/var/lib/geos", "city")
	require.ErrorContains(t, err, "empty path")

	_, err = OpenPatchFile(filepath.Join(dir, "missing.json"), "/var/lib/geos", "city")
	require.ErrorContains(t, err, "patch file")

	_, err = OpenPatchFile(`C:\geo\patch.json`, "/var/lib/geos", "city")
	require.Error(t, err)
	require.NotContains(t, err.Error(), "unsupported patch source")

	_, err = OpenPatchFile("ftp://example.com/city_patch.gz", "/var/lib/geos", "city")
	require.Error(t, err)
	require.NotContains(t, err.Error(), "unsupported patch source")
}

func TestReaderSkipsDownloadWithoutRemoteURL(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	missing := filepath.Join(dir, "no-such-dir", "patch.json")
	file := NewTSUpdatableFile(missing, "")

	_, err := file.Reader(context.Background())
	require.ErrorIs(t, err, ErrFileNotExists)

	_, statErr := os.Stat(filepath.Dir(missing))
	require.True(t, os.IsNotExist(statErr))
	_, statErr = os.Stat(missing + ".tmp")
	require.True(t, os.IsNotExist(statErr))
}
