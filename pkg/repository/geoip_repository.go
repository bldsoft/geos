package repository

import (
	"context"
	"log"
	"net"

	"github.com/bldsoft/geos/pkg/entity"
	"github.com/oschwald/maxminddb-golang"
)

type GeoIpRepository struct {
	db *maxminddb.Reader
}

func NewGeoIpRepository(file string) *GeoIpRepository {
	db, err := maxminddb.Open(file)
	if err != nil {
		log.Fatalf("Failed to open db: %s", err)
	}
	return &GeoIpRepository{db: db}
}

func lookup[T any](db *maxminddb.Reader, ip net.IP) (*T, error) {
	var obj T
	return &obj, db.Lookup(ip, &obj)
}

func (r *GeoIpRepository) Country(ctx context.Context, ip net.IP) (*entity.Country, error) {
	return lookup[entity.Country](r.db, ip)
}

func (r *GeoIpRepository) City(ctx context.Context, ip net.IP) (*entity.City, error) {
	return lookup[entity.City](r.db, ip)
}

func (r *GeoIpRepository) CityLite(ctx context.Context, ip net.IP, lang string) (*entity.CityLite, error) {
	cityLiteDb, err := lookup[entity.CityLiteDb](r.db, ip)
	if err != nil {
		return nil, err
	}
	return entity.DbToCityLite(cityLiteDb, lang), nil
}
