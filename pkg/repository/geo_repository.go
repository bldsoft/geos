package repository

import (
	"context"
	"log"
	"net"

	"github.com/oschwald/geoip2-golang"
)

type GeoRepository struct {
	db *geoip2.Reader
}

func NewGeoRepository(file string) *GeoRepository {
	db, err := geoip2.Open(file)
	if err != nil {
		log.Fatalf("Failed to open db: %s", err)
	}
	return &GeoRepository{db: db}
}

func (r *GeoRepository) Country(ctx context.Context, ip net.IP) (*geoip2.Country, error) {
	return r.db.Country(ip)
}

func (r *GeoRepository) City(ctx context.Context, ip net.IP) (*geoip2.City, error) {
	return r.db.City(ip)
}
