package test

import (
	"context"
	"testing"

	"github.com/bldsoft/geos/pkg/entity"
	"github.com/bldsoft/geos/pkg/storage"
	"github.com/stretchr/testify/assert"
)

func BenchmarkGeonamesCity(b *testing.B) {
	storage := storage.NewGeoNamesStorage(b.TempDir())
	storage.WaitReady()
	tests := []struct {
		Name   string
		Filter entity.GeoNameFilter
	}{
		{
			"search by name", entity.GeoNameFilter{
				NamePrefix: "Minsk",
			},
		},
		{
			"search by name prefix", entity.GeoNameFilter{
				NamePrefix: "Min",
			},
		},
		{
			"search by country", entity.GeoNameFilter{
				CountryCodes: []string{"BY"},
			},
		},
		{
			"search by name and country", entity.GeoNameFilter{
				CountryCodes: []string{"BY"},
				NamePrefix:   "Minsk",
			},
		},
		{
			"search by name prefix and country", entity.GeoNameFilter{
				CountryCodes: []string{"BY"},
				NamePrefix:   "Min",
			},
		},
	}
	for _, tt := range tests {
		b.Run(tt.Name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				city, err := storage.Cities(context.Background(), tt.Filter)
				assert.NotEmpty(b, city)
				assert.NoError(b, err)
			}
		})
	}
}
