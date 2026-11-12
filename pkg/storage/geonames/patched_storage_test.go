package geonames

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/bldsoft/geos/pkg/entity"
	"github.com/bldsoft/geos/pkg/storage/source"
	"github.com/stretchr/testify/require"
)

func TestPatchedStorageAddConsultsCustom(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	patchPath := filepath.Join(dir, "geonames_patch.json")
	body := `[
		{
			"city": {"geoNameID": 1, "name": "Testville"},
			"subdivision": {"geoNameID": 2, "name": "Testsub"},
			"country": {"geoNameID": 3, "name": "TC"},
			"continent": {"geoNameID": 4, "name": "EU"}
		}
	]`
	require.NoError(t, os.WriteFile(patchPath, []byte(body), 0o644))

	custom := NewCustomStorage(context.Background(), source.NewTSUpdatableFile(patchPath, ""))
	patched := &PatchedStorage{MultiStorage: *NewMultiStorage()}
	patched.Add(custom)

	cities, err := patched.Cities(context.Background(), entity.GeoNameFilter{})
	require.NoError(t, err)
	require.Len(t, cities, 1)
	require.Equal(t, "Testville", cities[0].Name)
}
