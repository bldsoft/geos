package source

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOpenPatchFile(t *testing.T) {
	t.Parallel()

	remote, err := OpenPatchFile("https://example.com/patches/city_patch.tar.gz", "/var/lib/geos", "city")
	require.NoError(t, err)
	require.Equal(t, "/var/lib/geos/city_patch.gz", remote.LocalPath)
	require.Equal(t, "https://example.com/patches/city_patch.tar.gz", remote.RemoteURL)

	local, err := OpenPatchFile("/home/setplex/W/Work/mmdbs/city_patch.gz", "/var/lib/geos", "city")
	require.NoError(t, err)
	require.Equal(t, "/home/setplex/W/Work/mmdbs/city_patch.gz", local.LocalPath)
	require.Empty(t, local.RemoteURL)

	fromFileURL, err := OpenPatchFile("file:///home/setplex/W/Work/mmdbs/city_patch.gz", "/var/lib/geos", "city")
	require.NoError(t, err)
	require.Equal(t, "/home/setplex/W/Work/mmdbs/city_patch.gz", fromFileURL.LocalPath)
	require.Empty(t, fromFileURL.RemoteURL)

	_, err = OpenPatchFile("ftp://example.com/city_patch.gz", "/var/lib/geos", "city")
	require.Error(t, err)
}
