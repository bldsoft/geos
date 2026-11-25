package repository

import (
	"context"
	"net"
	"net/netip"
	"os"
	"path/filepath"
	"testing"

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
