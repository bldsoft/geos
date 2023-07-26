package service

import (
	"context"
	"net"

	"github.com/bldsoft/geos/pkg/entity"
	"github.com/bldsoft/geos/pkg/microservice/middleware"
	"github.com/bldsoft/geos/pkg/repository"
)

type DumpFormat = repository.DumpFormat
type DBType = repository.MaxmindDBType

type GeoRepository interface {
	Country(ctx context.Context, ip net.IP) (*entity.Country, error)
	City(ctx context.Context, ip net.IP, includeISP bool) (*entity.City, error)
	CityLite(ctx context.Context, ip net.IP, lang string) (*entity.CityLite, error)
	Database(ctx context.Context, dbType DBType, format DumpFormat) (*entity.Database, error)
}

type GeoIpService struct {
	rep GeoRepository
}

func NewGeoIpService(rep GeoRepository) *GeoIpService {
	return &GeoIpService{rep: rep}
}

func (s *GeoIpService) ip(ctx context.Context, address string) (net.IP, error) {
	if address == "me" {
		address = middleware.GetRealIP(ctx)
	}

	// cut the port
	if host, _, err := net.SplitHostPort(address); err == nil {
		address = host
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

func (s *GeoIpService) City(ctx context.Context, address string, includeISP bool) (*entity.City, error) {
	ip, err := s.ip(ctx, address)
	if err != nil {
		return nil, err
	}
	return s.rep.City(ctx, ip, includeISP)
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

func (r *GeoIpService) Database(ctx context.Context, dbType DBType, format DumpFormat) (*entity.Database, error) {
	return r.rep.Database(ctx, dbType, format)
}
