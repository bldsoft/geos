package controller

import (
	"context"
	"net"

	"github.com/oschwald/geoip2-golang"
)

type GeoService interface {
	Country(ctx context.Context, ip net.IP) (*geoip2.Country, error)
	City(ctx context.Context, ip net.IP) (*geoip2.City, error)
}
