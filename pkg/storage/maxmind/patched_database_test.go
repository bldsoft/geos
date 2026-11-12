package maxmind

import (
	"context"
	"io"
	"net/netip"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/bldsoft/geos/pkg/entity"
	"github.com/bldsoft/geos/pkg/storage/source"
	"github.com/bldsoft/geos/pkg/utils"
	"github.com/oschwald/maxminddb-golang/v2"
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
	require.ErrorIs(t, err, utils.ErrNotFound)
}

func TestLocalPatchReloadsOnMtime(t *testing.T) {
	t.Parallel()

	v1 := `{
		"38.65.120.0/22": {
			"country": {"isoCode": "US", "names": {"en": "United States"}}
		}
	}`
	v2 := `{
		"38.65.120.0/22": {
			"country": {"isoCode": "VE", "names": {"en": "Venezuela"}}
		}
	}`

	dir := t.TempDir()
	patchPath := filepath.Join(dir, "city_patch.json")
	require.NoError(t, os.WriteFile(patchPath, []byte(v1), 0o644))
	past := time.Now().Add(-2 * time.Hour)
	require.NoError(t, os.Chtimes(patchPath, past, past))

	ctx := context.Background()
	custom := NewCustomDatabase(ctx, source.NewTSUpdatableFile(patchPath, ""))

	var got entity.Country
	err := custom.Lookup(ctx, netip.MustParseAddr("38.65.120.8"), &got)
	require.NoError(t, err)
	require.Equal(t, "US", got.Country.IsoCode)

	require.NoError(t, os.WriteFile(patchPath, []byte(v2), 0o644))
	future := time.Now().Add(2 * time.Hour)
	require.NoError(t, os.Chtimes(patchPath, future, future))

	upd, err := custom.CheckUpdates(ctx)
	require.NoError(t, err)
	require.Greater(t, upd.RemoteVersion.Compare(upd.CurrentVersion), 0)

	require.NoError(t, custom.Update(ctx, false))

	got = entity.Country{}
	err = custom.Lookup(ctx, netip.MustParseAddr("38.65.120.8"), &got)
	require.NoError(t, err)
	require.Equal(t, "VE", got.Country.IsoCode)
}

func TestMergedRawDataCachedUntilUpdate(t *testing.T) {
	t.Parallel()

	basePatch := mustDatabasePatch(t, `{
		"8.8.8.0/24": {
			"country": {"isoCode": "US", "names": {"en": "United States"}}
		}
	}`)
	base := &MaxmindDatabase{}
	raw := append([]byte(nil), basePatch.dbRaw...)
	base.reader.Store(basePatch.db)
	base.dbRaw.Store(&raw)

	dir := t.TempDir()
	patchPath := filepath.Join(dir, "city_patch.json")
	require.NoError(t, os.WriteFile(patchPath, []byte(`{
		"38.65.120.0/22": {
			"country": {"isoCode": "VE", "names": {"en": "Venezuela"}}
		}
	}`), 0o644))
	past := time.Now().Add(-2 * time.Hour)
	require.NoError(t, os.Chtimes(patchPath, past, past))

	ctx := context.Background()
	custom := NewCustomDatabase(ctx, source.NewTSUpdatableFile(patchPath, ""))
	db := NewPatchedDatabase(base).SetCustom(custom)

	before := db.mergeCount.Load()
	first, err := db.RawData(ctx)
	require.NoError(t, err)
	second, err := db.RawData(ctx)
	require.NoError(t, err)
	require.Equal(t, before+1, db.mergeCount.Load())

	firstBytes, err := io.ReadAll(first)
	require.NoError(t, err)
	secondBytes, err := io.ReadAll(second)
	require.NoError(t, err)
	require.Equal(t, firstBytes, secondBytes)

	require.NoError(t, os.WriteFile(patchPath, []byte(`{
		"38.65.120.0/22": {
			"country": {"isoCode": "US", "names": {"en": "United States"}}
		}
	}`), 0o644))
	future := time.Now().Add(2 * time.Hour)
	require.NoError(t, os.Chtimes(patchPath, future, future))
	require.NoError(t, custom.Update(ctx, false))
	db.invalidateMerged()

	third, err := db.RawData(ctx)
	require.NoError(t, err)
	require.Equal(t, before+2, db.mergeCount.Load())
	thirdBytes, err := io.ReadAll(third)
	require.NoError(t, err)
	require.NotEqual(t, firstBytes, thirdBytes)
}

func TestMergedBuildEpoch(t *testing.T) {
	t.Parallel()

	db := NewMultiMaxMindDB()
	require.Equal(t, uint(10), db.mergedBuildEpoch(10))
	require.Equal(t, uint(20), db.mergedBuildEpoch(10, 20))
	require.Equal(t, uint(11), db.mergedBuildEpoch(10, 5))
	require.Equal(t, uint(20), db.mergedBuildEpoch(10, 5, 20))
	require.Equal(t, uint(10), db.mergedBuildEpoch(10, 10))
}

func TestMergedDatabaseUsesBaseMetadata(t *testing.T) {
	t.Parallel()

	const (
		baseEpoch  = uint(1_700_000_000)
		olderPatch = int64(1_600_000_000)
		newerPatch = int64(1_800_000_000)
	)

	ctx := context.Background()
	db := patchedDBWithEpoch(t, time.Unix(olderPatch, 0), baseEpoch)

	meta, err := db.MetaData(ctx)
	require.NoError(t, err)
	require.Equal(t, "GeoIP2-City", meta.DatabaseType)
	require.Equal(t, []string{"en"}, meta.Languages)
	require.Equal(t, uint(6), meta.IPVersion)
	require.Equal(t, uint(28), meta.RecordSize)
	require.Equal(t, "GeoIP2 City database patched by GEOS service.", meta.Description["en"])
	require.Equal(t, baseEpoch+1, meta.BuildEpoch)

	fileMeta := metadataOf(t, db)
	require.Equal(t, meta.DatabaseType, fileMeta.DatabaseType)
	require.Equal(t, meta.Languages, fileMeta.Languages)
	require.Equal(t, meta.IPVersion, fileMeta.IPVersion)
	require.Equal(t, meta.RecordSize, fileMeta.RecordSize)
	require.Equal(t, meta.Description["en"], fileMeta.Description["en"])
	require.Equal(t, meta.BuildEpoch, fileMeta.BuildEpoch)

	newer := patchedDBWithEpoch(t, time.Unix(newerPatch, 0), baseEpoch)
	customMeta, err := newer.custom.MetaData(ctx)
	require.NoError(t, err)
	require.Equal(t, uint(newerPatch), customMeta.BuildEpoch)

	newerMeta, err := newer.MetaData(ctx)
	require.NoError(t, err)
	require.Equal(t, uint(newerPatch), newerMeta.BuildEpoch)
	require.Equal(t, uint(newerPatch), metadataOf(t, newer).BuildEpoch)
}

func patchedDBWithEpoch(t *testing.T, patchTime time.Time, baseEpoch uint) *PatchedDatabase {
	t.Helper()

	basePatch := mustDatabasePatch(t, `{
		"8.8.8.0/24": {
			"country": {"isoCode": "US", "names": {"en": "United States"}}
		}
	}`)
	basePatch.WithMetadata(maxminddb.Metadata{
		DatabaseType:             "GeoIP2-City",
		Description:              map[string]string{"en": "GeoIP2 City database"},
		Languages:                []string{"en"},
		BinaryFormatMajorVersion: 2,
		IPVersion:                6,
		RecordSize:               28,
		BuildEpoch:               baseEpoch,
		NodeCount:                basePatch.db.Metadata.NodeCount,
	})
	base := &MaxmindDatabase{}
	raw := append([]byte(nil), basePatch.dbRaw...)
	base.reader.Store(basePatch.db)
	base.dbRaw.Store(&raw)

	dir := t.TempDir()
	patchPath := filepath.Join(dir, "city_patch.json")
	require.NoError(t, os.WriteFile(patchPath, []byte(`{
		"38.65.120.0/22": {
			"country": {"isoCode": "VE", "names": {"en": "Venezuela"}}
		}
	}`), 0o644))
	require.NoError(t, os.Chtimes(patchPath, patchTime, patchTime))

	custom := NewCustomDatabase(context.Background(), source.NewTSUpdatableFile(patchPath, ""))
	return NewPatchedDatabase(base).SetCustom(custom)
}

func metadataOf(t *testing.T, db *PatchedDatabase) maxminddb.Metadata {
	t.Helper()

	raw, err := db.RawData(context.Background())
	require.NoError(t, err)
	data, err := io.ReadAll(raw)
	require.NoError(t, err)
	reader, err := maxminddb.OpenBytes(data)
	require.NoError(t, err)
	return reader.Metadata
}

func mustDatabasePatch(t *testing.T, body string) *DatabasePatch {
	t.Helper()
	reader, err := NewJSONRecordReader(strings.NewReader(body))
	require.NoError(t, err)
	patch, err := NewDatabasePatch(reader)
	require.NoError(t, err)
	return patch
}
