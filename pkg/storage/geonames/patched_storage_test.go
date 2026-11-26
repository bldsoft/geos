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

func TestStoragePatchHonorsLimit(t *testing.T) {
	t.Parallel()

	var records []CustomGeonamesRecord
	for i := 1; i <= 5; i++ {
		records = append(records, NewCustomGeonamesRecord("Place", i))
	}
	patch := NewStoragePatch(records)
	ctx := context.Background()

	cities, err := patch.Cities(ctx, entity.GeoNameFilter{Limit: 2})
	require.NoError(t, err)
	require.Len(t, cities, 2)

	subdivisions, err := patch.Subdivisions(ctx, entity.GeoNameFilter{Limit: 3})
	require.NoError(t, err)
	require.Len(t, subdivisions, 3)

	countries, err := patch.Countries(ctx, entity.GeoNameFilter{Limit: 1})
	require.NoError(t, err)
	require.Len(t, countries, 1)

	cities, err = patch.Cities(ctx, entity.GeoNameFilter{})
	require.NoError(t, err)
	require.Len(t, cities, 5)
}

func TestMultiStorageStopsWhenLimitIsFilled(t *testing.T) {
	t.Parallel()

	newPatch := func(count int) Storage {
		var records []CustomGeonamesRecord
		for i := 1; i <= count; i++ {
			records = append(records, NewCustomGeonamesRecord("Place", i))
		}

		return NewStoragePatch(records)
	}
	multi := NewMultiStorage(newPatch(3), newPatch(5))
	ctx := context.Background()

	tests := []struct {
		limit uint32
		want  int
	}{
		{limit: 0, want: 8},
		{limit: 2, want: 2},
		{limit: 3, want: 3},
		{limit: 4, want: 4},
		{limit: 8, want: 8},
		{limit: 20, want: 8},
	}
	for _, tt := range tests {
		filter := entity.GeoNameFilter{Limit: tt.limit}

		cities, err := multi.Cities(ctx, filter)
		require.NoError(t, err)
		require.Len(t, cities, tt.want, "cities, limit %d", tt.limit)

		subdivisions, err := multi.Subdivisions(ctx, filter)
		require.NoError(t, err)
		require.Len(t, subdivisions, tt.want, "subdivisions, limit %d", tt.limit)

		countries, err := multi.Countries(ctx, filter)
		require.NoError(t, err)
		require.Len(t, countries, tt.want, "countries, limit %d", tt.limit)
	}
}
