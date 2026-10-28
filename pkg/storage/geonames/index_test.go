package geonames

import (
	"testing"

	"github.com/bldsoft/geos/pkg/entity"
	"github.com/mkrou/geonames/models"
	"github.com/stretchr/testify/require"
)

func TestIndexNamePrefix(t *testing.T) {
	cities := []*entity.GeoName{
		{Geoname: &models.Geoname{Id: 1, Name: "Minsk", CountryCode: "BY"}},
		{Geoname: &models.Geoname{Id: 2, Name: "Minden", CountryCode: "DE"}},
		{Geoname: &models.Geoname{Id: 3, Name: "Moscow", CountryCode: "RU"}},
		{Geoname: &models.Geoname{Id: 4, Name: "Paris", CountryCode: "FR"}},
		{Geoname: &models.Geoname{Id: 5, Name: "paris", CountryCode: "US"}},
	}

	var idx index[*entity.GeoName]
	idx.Init(cities)

	got := idx.GetFiltered(entity.GeoNameFilter{NamePrefix: "Min"})
	require.Len(t, got, 2)
	require.Equal(t, "Minden", got[0].GetName())
	require.Equal(t, "Minsk", got[1].GetName())

	got = idx.GetFiltered(entity.GeoNameFilter{NamePrefix: "par"})
	require.Len(t, got, 2)
	require.Equal(t, 4, got[0].GetGeoNameID())
	require.Equal(t, 5, got[1].GetGeoNameID())

	got = idx.GetFiltered(entity.GeoNameFilter{
		NamePrefix:   "Min",
		CountryCodes: []string{"BY"},
	})
	require.Len(t, got, 1)
	require.Equal(t, "Minsk", got[0].GetName())
}
