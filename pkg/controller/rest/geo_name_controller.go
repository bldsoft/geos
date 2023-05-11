package rest

import (
	"net/http"

	"github.com/bldsoft/geos/pkg/controller"
	"github.com/bldsoft/geos/pkg/entity"
	gost "github.com/bldsoft/gost/controller"
	"github.com/bldsoft/gost/log"
)

type GeoNameController struct {
	gost.BaseController
	geoNameService controller.GeoNameService
}

func NewGeoNameController(geoNameService controller.GeoNameService) *GeoNameController {
	return &GeoNameController{geoNameService: geoNameService}
}

func (c *GeoNameController) getGeoNameFilter(r *http.Request) entity.GeoNameFilter {
	codes, _ := gost.GetQueryOptionSlice[string](r, "country-codes")
	namePrefix, _ := gost.GetQueryOption(r, "name-prefix", "")
	geoNameIDs, _ := gost.GetQueryOptionSlice[uint32](r, "geoname-ids")
	limit, _ := gost.GetQueryOption(r, "limit", uint32(0))
	return entity.GeoNameFilter{
		CountryCodes: codes,
		NamePrefix:   namePrefix,
		GeoNameIDs:   geoNameIDs,
		Limit:        limit,
	}
}

// @Summary continent
// @Produce json
// @Tags geonames
// @Success 200 {object} []entity.geoNameContinentJson
// @Router /geoname/continent [get]
func (c *GeoNameController) GetGeoNameContinentsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	countries := c.geoNameService.Continents(ctx)
	c.ResponseJson(w, r, countries, false)
}

// @Summary country
// @Produce json
// @Tags geonames
// @Param country-codes query []string false "comma separated list of country codes"
// @Param name-prefix query string false "name prefix"
// @Param geoname-ids query []integer false "comma separated list of GeoNames ids"
// @Success 200 {object} []entity.GeoNameCountry
// @Failure 400 {string} string "error"
// @Failure 500 {string} string "error"
// @Router /geoname/country [get]
func (c *GeoNameController) GetGeoNameCountriesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	filter := c.getGeoNameFilter(r)
	countries, err := c.geoNameService.Countries(ctx, filter)
	if err != nil {
		log.FromContext(ctx).Error(err.Error())
		c.ResponseError(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	c.ResponseJson(w, r, countries, false)
}

// @Summary city lite
// @Produce json
// @Tags geonames
// @Param country-codes query []string false "comma separated list of country codes"
// @Param name-prefix query string false "name prefix"
// @Param geoname-ids query []integer false "comma separated list of GeoNames ids"
// @Success 200 {object} []entity.GeoNameAdminSubdivision
// @Failure 400 {string} string "error"
// @Failure 500 {string} string "error"
// @Router /geoname/subdivision [get]
func (c *GeoNameController) GetGeoNameSubdivisionsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	filter := c.getGeoNameFilter(r)
	subdivisions, err := c.geoNameService.Subdivisions(ctx, filter)
	if err != nil {
		log.FromContext(ctx).Error(err.Error())
		c.ResponseError(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	c.ResponseJson(w, r, subdivisions, false)
}

// @Summary city
// @Produce json
// @Tags geonames
// @Param country-codes query []string false "comma separated list of country codes"
// @Param name-prefix query string false "name prefix"
// @Param geoname-ids query []integer false "comma separated list of GeoNames ids"
// @Success 200 {object} []entity.GeoName
// @Failure 400 {string} string "error"
// @Failure 500 {string} string "error"
// @Router /geoname/city [get]
func (c *GeoNameController) GetGeoNameCitiesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	filter := c.getGeoNameFilter(r)
	cities, err := c.geoNameService.Cities(ctx, filter)
	if err != nil {
		log.FromContext(ctx).Error(err.Error())
		c.ResponseError(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	c.ResponseJson(w, r, cities, false)
}
