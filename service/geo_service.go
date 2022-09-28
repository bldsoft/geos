package service

import (
	"context"
	"net"

	"github.com/oschwald/geoip2-golang"
)

type GeoRepository interface {
	Country(ctx context.Context, ip net.IP) (*geoip2.Country, error)
	City(ctx context.Context, ip net.IP) (*geoip2.City, error)
}

type GeoService struct {
	rep GeoRepository
}

func NewGeoService(rep GeoRepository) *GeoService {
	return &GeoService{rep: rep}
}

func (s *GeoService) Country(ctx context.Context, ip net.IP) (*geoip2.Country, error) {
	return s.rep.Country(ctx, ip)
}

func (s *GeoService) City(ctx context.Context, ip net.IP) (*geoip2.City, error) {
	return s.rep.City(ctx, ip)
}
