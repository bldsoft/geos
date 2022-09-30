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

type GeoIpService struct {
	rep GeoRepository
}

func NewGeoService(rep GeoRepository) *GeoIpService {
	return &GeoIpService{rep: rep}
}

func (s *GeoIpService) Country(ctx context.Context, ip net.IP) (*geoip2.Country, error) {
	return s.rep.Country(ctx, ip)
}

func (s *GeoIpService) City(ctx context.Context, ip net.IP) (*geoip2.City, error) {
	return s.rep.City(ctx, ip)
}
