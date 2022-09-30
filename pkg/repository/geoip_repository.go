package repository

import (
	"context"
	"log"
	"net"

	"github.com/oschwald/geoip2-golang"
)

type GeoIpRepository struct {
	db *geoip2.Reader
}

func NewGeoIpRepository(file string) *GeoIpRepository {
	db, err := geoip2.Open(file)
	if err != nil {
		log.Fatalf("Failed to open db: %s", err)
	}
	return &GeoIpRepository{db: db}
}

func (r *GeoIpRepository) Country(ctx context.Context, ip net.IP) (*geoip2.Country, error) {
	return r.db.Country(ip)
}

func (r *GeoIpRepository) City(ctx context.Context, ip net.IP) (*geoip2.City, error) {
	return r.db.City(ip)
}
