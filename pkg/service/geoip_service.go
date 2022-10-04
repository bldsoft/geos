package service

import (
	"context"
	"net"

	"github.com/bldsoft/geos/pkg/entity"
	"github.com/bldsoft/geos/pkg/microservice/middleware"
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

func (s *GeoIpService) ip(ctx context.Context, address string) (net.IP, error) {
	if address == "me" {
		address = middleware.GetRealIP(ctx)
	}

	if ip := net.ParseIP(address); ip != nil {
		return ip, nil
	}
	ips, err := net.LookupIP(address)
	if err != nil {
		return nil, err
	}
	return ips[0], nil
}

func (s *GeoIpService) Country(ctx context.Context, address string) (*entity.Country, error) {
	ip, err := s.ip(ctx, address)
	if err != nil {
		return nil, err
	}
	return s.rep.Country(ctx, ip)
}

func (s *GeoIpService) City(ctx context.Context, address string) (*entity.City, error) {
	ip, err := s.ip(ctx, address)
	if err != nil {
		return nil, err
	}
	return s.rep.City(ctx, ip)
}

func (s *GeoIpService) CityLite(ctx context.Context, address string, lang string) (*entity.CityLite, error) {
	ip, err := s.ip(ctx, address)
	if err != nil {
		return nil, err
	}

	if len(lang) == 0 {
		lang = "en"
	}

	return s.rep.CityLite(ctx, ip, lang)
}
