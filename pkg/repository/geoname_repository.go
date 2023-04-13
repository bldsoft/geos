package repository

import (
	"context"

	"github.com/bldsoft/geos/pkg/entity"
	"github.com/mkrou/geonames"
	"github.com/mkrou/geonames/models"
)

type GeoNameRepository struct {
	parser geonames.Parser
}

func NewGeoNamesRepository() *GeoNameRepository {
	return &GeoNameRepository{parser: geonames.NewParser()}
}

func (r *GeoNameRepository) Countries(ctx context.Context) ([]*entity.GeoNameCountry, error) {
	var res []*entity.GeoNameCountry
	r.parser.GetCountries(func(c *models.Country) error {
		res = append(res, (*entity.GeoNameCountry)(c))
		return nil
	})
	return res, nil
}

func (r *GeoNameRepository) Subdivisions(ctx context.Context) ([]*entity.AdminSubdivision, error) {
	var res []*entity.AdminSubdivision
	r.parser.GetAdminSubdivisions(func(s *models.AdminSubdivision) error {
		res = append(res, (*entity.AdminSubdivision)(s))
		return nil
	})
	return res, nil
}

func (r *GeoNameRepository) Cities(ctx context.Context) ([]*entity.Geoname, error) {
	var res []*entity.Geoname
	r.parser.GetGeonames(geonames.Cities500, func(g *models.Geoname) error {
		res = append(res, (*entity.Geoname)(g))
		return nil
	})
	return res, nil
}
