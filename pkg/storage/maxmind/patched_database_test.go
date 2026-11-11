package maxmind

import (
	"context"
	"net/netip"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bldsoft/geos/pkg/entity"
	"github.com/bldsoft/geos/pkg/storage/source"
	"github.com/stretchr/testify/require"
)

func TestPatchOverridesBaseCity(t *testing.T) {
	baseJSON := `{
		"38.65.120.0/22": {
			"country": {"isoCode": "US", "names": {"en": "United States"}},
			"city": {"names": {"en": "Marble Hill"}}
		},
		"8.8.8.0/24": {
			"country": {"isoCode": "US", "names": {"en": "United States"}},
			"city": {"names": {"en": "Mountain View"}}
		}
	}`
	patchJSON := `{
		"38.65.120.0/22": {
			"country": {"isoCode": "VE", "names": {"en": "Venezuela"}},
			"city": {"names": {"en": "La Villa del Rosario"}}
		}
	}`

	basePatch := mustDatabasePatch(t, baseJSON)
	base := &MaxmindDatabase{}
	raw := append([]byte(nil), basePatch.dbRaw...)
	base.reader.Store(basePatch.db)
	base.dbRaw.Store(&raw)

	dir := t.TempDir()
	patchPath := filepath.Join(dir, "city_patch.json")
	require.NoError(t, os.WriteFile(patchPath, []byte(patchJSON), 0o644))
	custom := NewCustomDatabase(context.Background(), source.NewTSUpdatableFile(patchPath, "https://example.com/city_patch.json"))

	db := NewPatchedDatabase(base).SetCustom(custom)

	var patched entity.Country
	err := db.Lookup(context.Background(), netip.MustParseAddr("38.65.120.8"), &patched)
	require.NoError(t, err)
	require.Equal(t, "VE", patched.Country.IsoCode)
	require.Equal(t, "Venezuela", patched.Country.Names["en"])

	var untouched entity.Country
	err = db.Lookup(context.Background(), netip.MustParseAddr("8.8.8.8"), &untouched)
	require.NoError(t, err)
	require.Equal(t, "US", untouched.Country.IsoCode)
	require.Equal(t, "United States", untouched.Country.Names["en"])

	var missing entity.Country
	err = db.Lookup(context.Background(), netip.MustParseAddr("1.1.1.1"), &missing)
	require.NoError(t, err)
	require.Empty(t, missing.Country.IsoCode)
}

func mustDatabasePatch(t *testing.T, body string) *DatabasePatch {
	t.Helper()
	reader, err := NewJSONRecordReader(strings.NewReader(body))
	require.NoError(t, err)
	patch, err := NewDatabasePatch(reader)
	require.NoError(t, err)
	return patch
}
