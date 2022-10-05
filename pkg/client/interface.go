package client

import (
	"context"

	"github.com/bldsoft/geos/pkg/entity"
)

type GeoIPClient interface {
	Country(ctx context.Context, address string) (*entity.Country, error)
	City(ctx context.Context, address string) (*entity.City, error)
	CityLite(ctx context.Context, address, lang string) (*entity.CityLite, error)
}
