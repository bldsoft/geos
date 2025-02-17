package rest

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/bldsoft/geos/pkg/controller"
	"github.com/bldsoft/geos/pkg/entity"
	"github.com/bldsoft/geos/pkg/service"
	"github.com/bldsoft/geos/pkg/utils"
	gost "github.com/bldsoft/gost/controller"
	"github.com/bldsoft/gost/log"
	gostUtils "github.com/bldsoft/gost/utils"
)

type GeoNameController struct {
	gost.BaseController
	geoNameService controller.GeoNameService
}

func NewGeoNameController(geoNameService controller.GeoNameService) *GeoNameController {
	return &GeoNameController{geoNameService: geoNameService}
}

func (c *GeoNameController) getGeoNameFilter(r *http.Request) (filter *entity.GeoNameFilter, err error) {
	filter = &entity.GeoNameFilter{}
	if r.Method == http.MethodPost {
		err = json.NewDecoder(r.Body).Decode(filter)
		return
	}

	filter, err = gostUtils.FromRequest[entity.GeoNameFilter](r)
	return
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
// @Failure 503 {string} string "error"
// @Router /geoname/country [get]
func (c *GeoNameController) GetGeoNameCountriesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	filter, err := c.getGeoNameFilter(r)
	if err != nil {
		c.responseError(w, r, err)
		return
	}

	countries, err := c.geoNameService.Countries(ctx, *filter)
	if err != nil {
		c.responseError(w, r, err)
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
// @Failure 503 {string} string "error"
// @Router /geoname/subdivision [get]
func (c *GeoNameController) GetGeoNameSubdivisionsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	filter, err := c.getGeoNameFilter(r)
	if err != nil {
		c.responseError(w, r, err)
		return
	}

	subdivisions, err := c.geoNameService.Subdivisions(ctx, *filter)
	if err != nil {
		c.responseError(w, r, err)
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
// @Failure 503 {string} string "error"
// @Router /geoname/city [get]
func (c *GeoNameController) GetGeoNameCitiesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	filter, err := c.getGeoNameFilter(r)
	if err != nil {
		c.responseError(w, r, err)
		return
	}

	cities, err := c.geoNameService.Cities(ctx, *filter)
	if err != nil {
		c.responseError(w, r, err)
		return
	}
	c.ResponseJson(w, r, cities, false)
}

// @Summary geonames csv dump
// @Produce text/csv
// @Tags geonames
// @Param format query string false "format" "csvWithNames"
// @Success 200 {object} string
// @Failure 400 {string} string "error"
// @Failure 500 {string} string "error"
// @Failure 503 {string} string "error"
// @Router /geoname/dump [get]
func (c *GeoNameController) GetDumpHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	format, _ := gost.GetQueryOption[string](r, "format")
	dump, err := c.geoNameService.Dump(ctx, service.DumpFormat(format))
	if err != nil {
		c.responseError(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(dump)
}

func (c *GeoNameController) responseError(w http.ResponseWriter, r *http.Request, err error) {
	log.FromContext(r.Context()).Error(err.Error())
	switch {
	case errors.Is(err, utils.ErrDisabled):
		c.ResponseError(w, err.Error(), http.StatusInternalServerError)
	case errors.Is(err, utils.ErrNotReady):
		c.ResponseError(w, err.Error(), http.StatusServiceUnavailable)
	case errors.Is(err, utils.ErrNotAvailable):
		c.ResponseError(w, err.Error(), http.StatusInternalServerError)
	default:
		c.ResponseError(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}
}
