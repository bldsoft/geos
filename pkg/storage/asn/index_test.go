package asn

import (
	"testing"

	"github.com/bldsoft/geos/pkg/entity"
	"github.com/stretchr/testify/require"
)

func testIndex() index {
	var idx index
	idx.Init([]*entity.ASN{
		{ASN: 46562, ASO: "Performive LLC"},
		{ASN: 15169, ASO: "Google LLC"},
		{ASN: 13335, ASO: "Cloudflare, Inc."},
	})
	return idx
}

func TestIndexNamePrefix(t *testing.T) {
	idx := testIndex()

	got := idx.GetFiltered(entity.ASNFilter{NamePrefix: "perf"})
	require.Len(t, got, 1)
	require.Equal(t, uint32(46562), got[0].ASN)

	got = idx.GetFiltered(entity.ASNFilter{NamePrefix: "AS15169"})
	require.Len(t, got, 1)
	require.Equal(t, uint32(15169), got[0].ASN)

	got = idx.GetFiltered(entity.ASNFilter{NamePrefix: "13335"})
	require.Len(t, got, 1)
	require.Equal(t, uint32(13335), got[0].ASN)
}

func TestIndexNumbersAndLimit(t *testing.T) {
	idx := testIndex()

	got := idx.GetFiltered(entity.ASNFilter{Numbers: []uint32{15169, 1}})
	require.Len(t, got, 1)
	require.Equal(t, uint32(15169), got[0].ASN)

	got = idx.GetFiltered(entity.ASNFilter{})
	require.GreaterOrEqual(t, len(got), 2)
	require.Equal(t, uint32(13335), got[0].ASN)
	require.Equal(t, uint32(15169), got[1].ASN)
}
