package controller

import (
	"context"
	"net"

	"github.com/bldsoft/geos/pkg/entity"
)

type GeoIpService interface {
	Country(ctx context.Context, ip net.IP) (*entity.Country, error)
	City(ctx context.Context, ip net.IP) (*entity.City, error)
	CityLite(ctx context.Context, ip net.IP, lang string) (*entity.CityLite, error)
}
