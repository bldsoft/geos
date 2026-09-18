package service

import (
	"context"
	"fmt"
	"net"
	"net/netip"

	"github.com/bldsoft/geos/pkg/entity"
	"github.com/bldsoft/geos/pkg/microservice/middleware"
	"github.com/bldsoft/geos/pkg/repository"
)

type DumpFormat = repository.DumpFormat
type DBType = repository.MaxmindDBType

type GeoRepository interface {
	Country(ctx context.Context, ip netip.Addr) (*entity.Country, error)
	City(ctx context.Context, ip netip.Addr, includeISP bool) (*entity.City, error)
	CityLite(ctx context.Context, ip netip.Addr, lang string) (*entity.CityLite, error)
	Hosting(ctx context.Context, ip netip.Addr) (*entity.Hosting, error)
	MetaData(ctx context.Context, dbType DBType) (*entity.MetaData, error)
	Database(ctx context.Context, dbType DBType, format DumpFormat) (*entity.Database, error)

	StartUpdate(ctx context.Context, dbType DBType) error
	CheckUpdates(ctx context.Context, dbType DBType) (entity.DBUpdate[entity.PatchedMMDBVersion], error)
	CurrentVersion(ctx context.Context, dbType DBType) (entity.PatchedMMDBVersion, error)
}

type GeoIpService struct {
	rep GeoRepository
}

func NewGeoIpService(rep GeoRepository) *GeoIpService {
	return &GeoIpService{rep: rep}
}

func (s *GeoIpService) ip(ctx context.Context, address string) (netip.Addr, error) {
	if address == "me" {
		address = middleware.GetRealIP(ctx)
	}

	// cut the port
	if host, _, err := net.SplitHostPort(address); err == nil {
		address = host
	}

	if ip, err := netip.ParseAddr(address); err == nil {
		return ip.Unmap(), nil
	}
	ips, err := net.DefaultResolver.LookupNetIP(ctx, "ip", address)
	if err != nil {
		return netip.Addr{}, err
	}
	if len(ips) == 0 {
		return netip.Addr{}, fmt.Errorf("no IP addresses found for %q", address)
	}
	return ips[0].Unmap(), nil
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

func (s *GeoIpService) Hosting(ctx context.Context, address string) (*entity.Hosting, error) {
	ip, err := s.ip(ctx, address)
	if err != nil {
		return nil, err
	}
	return s.rep.Hosting(ctx, ip)
}

func (r *GeoIpService) MetaData(ctx context.Context, dbType DBType) (*entity.MetaData, error) {
	return r.rep.MetaData(ctx, dbType)
}

func (r *GeoIpService) Database(ctx context.Context, dbType DBType, format DumpFormat) (*entity.Database, error) {
	return r.rep.Database(ctx, dbType, format)
}

func (r *GeoIpService) CheckUpdates(ctx context.Context, dbType DBType) (entity.DBUpdate[entity.PatchedMMDBVersion], error) {
	return r.rep.CheckUpdates(ctx, dbType)
}

func (r *GeoIpService) CurrentVersion(ctx context.Context, dbType DBType) (entity.PatchedMMDBVersion, error) {
	return r.rep.CurrentVersion(ctx, dbType)
}

func (r *GeoIpService) StartUpdate(ctx context.Context, dbType DBType) error {
	return r.rep.StartUpdate(ctx, dbType)
}
