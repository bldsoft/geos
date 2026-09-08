package asn

import (
	"context"
	"testing"

	"github.com/bldsoft/geos/pkg/entity"
	"github.com/bldsoft/geos/pkg/utils"
	"github.com/stretchr/testify/require"
)

func TestGetFilteredLimitAndNotReady(t *testing.T) {
	s := &ASNStorage{}
	_, err := s.GetFiltered(context.Background(), entity.ASNFilter{})
	require.ErrorIs(t, err, utils.ErrNotReady)

	idx := testIndex()
	s.catalog.Store(&idx)
	got, err := s.GetFiltered(context.Background(), entity.ASNFilter{Limit: 2})
	require.NoError(t, err)
	require.Len(t, got, 2)
	require.Equal(t, uint32(13335), got[0].ASN)
	require.Equal(t, uint32(15169), got[1].ASN)
}
