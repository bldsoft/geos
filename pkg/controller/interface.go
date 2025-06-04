package controller

import (
	"context"

	"github.com/bldsoft/geos/pkg/entity"
	"github.com/bldsoft/geos/pkg/service"
)

type GeoIpService interface {
	Country(ctx context.Context, address string) (*entity.Country, error)
	City(ctx context.Context, address string, includeISP bool) (*entity.City, error)
	CityLite(ctx context.Context, address string, lang string) (*entity.CityLite, error)
	MetaData(ctx context.Context, dbType service.DBType) (*entity.MetaData, error)
	Database(ctx context.Context, dbType service.DBType, format service.DumpFormat) (*entity.Database, error)
	CheckUpdates(ctx context.Context) (bool, error)
	Download(ctx context.Context, update ...bool) error
}

type GeoNameService interface {
	Continents(ctx context.Context) []*entity.GeoNameContinent
	Countries(ctx context.Context, filter entity.GeoNameFilter) ([]*entity.GeoNameCountry, error)
	Subdivisions(ctx context.Context, filter entity.GeoNameFilter) ([]*entity.GeoNameAdminSubdivision, error)
	Cities(ctx context.Context, filter entity.GeoNameFilter) ([]*entity.GeoName, error)
	Dump(ctx context.Context, format service.DumpFormat) ([]byte, error)
	CheckUpdates(ctx context.Context) (bool, error)
	Download(ctx context.Context, update ...bool) error
}
