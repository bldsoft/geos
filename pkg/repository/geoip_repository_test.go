package repository

import (
	"context"
	"net"
	"net/netip"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/maxmind/mmdbwriter"
	"github.com/maxmind/mmdbwriter/mmdbtype"
	"github.com/stretchr/testify/require"

	"github.com/bldsoft/geos/pkg/entity"
)

func writeTestISPDB(t *testing.T, path string) {
	t.Helper()

	tree, err := mmdbwriter.New(mmdbwriter.Options{DatabaseType: "GeoIP2-ISP", IncludeReservedNetworks: true})
	require.NoError(t, err)
	_, network, err := net.ParseCIDR("203.0.113.0/24")
	require.NoError(t, err)
	require.NoError(t, tree.Insert(network, mmdbtype.Map{
		"isp":                      mmdbtype.String("test isp"),
		"autonomous_system_number": mmdbtype.Uint32(64500),
	}))

	f, err := os.Create(path)
	require.NoError(t, err)
	defer f.Close()
	_, err = tree.WriteTo(f)
	require.NoError(t, err)
}

func TestOpenPatchedDBMissingPatchKeepsOptionalDBEnabled(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "isp.mmdb")
	writeTestISPDB(t, dbPath)

	db := openPatchedDB[entity.ISP](DBConfig{
		LocalPath:       dbPath,
		PatchesLocation: filepath.Join(dir, "mistyped_patch.tar.gz"),
	}, string(MaxmindDBTypeISP), dir, false)

	require.NotNil(t, db)
	var isp entity.ISP
	require.NoError(t, db.Lookup(context.Background(), netip.MustParseAddr("203.0.113.7"), &isp))
	require.Equal(t, "test isp", isp.ISP)
	require.EqualValues(t, 64500, isp.AutonomousSystemNumber)
}

func writeTestCountryDB(t *testing.T, path string) {
	t.Helper()

	tree, err := mmdbwriter.New(mmdbwriter.Options{DatabaseType: "GeoIP2-City", IncludeReservedNetworks: true})
	require.NoError(t, err)
	_, network, err := net.ParseCIDR("203.0.113.0/24")
	require.NoError(t, err)
	require.NoError(t, tree.Insert(network, mmdbtype.Map{
		"country": mmdbtype.Map{"iso_code": mmdbtype.String("US")},
	}))

	f, err := os.Create(path)
	require.NoError(t, err)
	defer f.Close()
	_, err = tree.WriteTo(f)
	require.NoError(t, err)
}

func TestMaxmindDBWithCachedCSVDumpUpdateReloadsLocalPatch(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "country.mmdb")
	writeTestCountryDB(t, dbPath)

	patchPath := filepath.Join(dir, "country_patch.json")
	writePatch := func(isoCode string, modTime time.Time) {
		t.Helper()
		body := `{"203.0.113.0/24": {"country": {"isoCode": "` + isoCode + `"}}}`
		require.NoError(t, os.WriteFile(patchPath, []byte(body), 0o644))
		require.NoError(t, os.Chtimes(patchPath, modTime, modTime))
	}
	writePatch("DE", time.Now().Add(-2*time.Hour))

	db := openPatchedDB[entity.City](DBConfig{
		LocalPath:       dbPath,
		PatchesLocation: patchPath,
	}, "country", dir, false)
	require.NotNil(t, db)

	ctx := context.Background()
	require.Eventually(t, func() bool {
		_, err := db.CSV(ctx, true)

		return err == nil
	}, 5*time.Second, 10*time.Millisecond)

	addr := netip.MustParseAddr("203.0.113.7")
	var country entity.City
	require.NoError(t, db.Lookup(ctx, addr, &country))
	require.Equal(t, "DE", country.Country.IsoCode)

	writePatch("FR", time.Now().Add(2*time.Hour))
	require.NoError(t, db.Update(ctx, false))

	country = entity.City{}
	require.NoError(t, db.Lookup(ctx, addr, &country))
	require.Equal(t, "FR", country.Country.IsoCode)
}
