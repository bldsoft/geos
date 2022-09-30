package service

import (
	"context"
	"net"

	"github.com/bldsoft/geos/pkg/entity"
)

type GeoRepository interface {
	Country(ctx context.Context, ip net.IP) (*entity.Country, error)
	City(ctx context.Context, ip net.IP) (*entity.City, error)
	CityLite(ctx context.Context, ip net.IP, lang string) (*entity.CityLite, error)
}

type GeoIpService struct {
	rep GeoRepository
}

func NewGeoService(rep GeoRepository) *GeoIpService {
	return &GeoIpService{rep: rep}
}

func (s *GeoIpService) Country(ctx context.Context, ip net.IP) (*entity.Country, error) {
	return s.rep.Country(ctx, ip)
}

func (s *GeoIpService) City(ctx context.Context, ip net.IP) (*entity.City, error) {
	return s.rep.City(ctx, ip)
}

func (s *GeoIpService) CityLite(ctx context.Context, ip net.IP, lang string) (*entity.CityLite, error) {
	return s.rep.CityLite(ctx, ip, lang)
}
