package rest

import (
	"net/http"

	"github.com/bldsoft/geos/pkg/controller"
	"github.com/bldsoft/geos/pkg/entity"
	_ "github.com/bldsoft/geos/pkg/entity"
	gost "github.com/bldsoft/gost/controller"
	"github.com/bldsoft/gost/log"
	"github.com/go-chi/chi/v5"
)

type GeoIpController struct {
	gost.BaseController
	geoIpService   controller.GeoIpService
	geoNameService controller.GeoNameService
}

func NewGeoIpController(geoIpService controller.GeoIpService, geoNameService controller.GeoNameService) (c *GeoIpController) {
	return &GeoIpController{geoIpService: geoIpService, geoNameService: geoNameService}
}

func (c *GeoIpController) address(r *http.Request) string {
	return chi.URLParam(r, "addr")
}

// @Summary city
// @Produce json
// @Tags geo IP
// @Param addr path string true "ip or hostname"
// @Success 200 {object} entity.City
// @Failure 400 {string} string "error"
// @Failure 500 {string} string "error"
// @Router /city/{addr} [get]
func (c *GeoIpController) GetCityHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	city, err := c.geoIpService.City(ctx, c.address(r))
	if err != nil {
		log.FromContext(ctx).Error(err.Error())
		c.ResponseError(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	c.ResponseJson(w, r, city)
}

// @Summary country
// @Produce json
// @Tags geo IP
// @Param addr path string true "ip or hostname"
// @Success 200 {object} entity.Country
// @Failure 400 {string} string "error"
// @Failure 500 {string} string "error"
// @Router /country/{addr} [get]
func (c *GeoIpController) GetCountryHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	country, err := c.geoIpService.Country(ctx, c.address(r))
	if err != nil {
		log.FromContext(ctx).Error(err.Error())
		c.ResponseError(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	c.ResponseJson(w, r, country)
}

// @Summary city lite
// @Produce json
// @Tags geo IP
// @Param addr path string true "ip or hostname"
// @Success 200 {object} entity.CityLite
// @Failure 400 {string} string "error"
// @Failure 500 {string} string "error"
// @Router /city-lite/{addr} [get]
func (c *GeoIpController) GetCityLiteHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	lang, _ := gost.GetQueryOption[string](r, "lang")
	city, err := c.geoIpService.CityLite(ctx, c.address(r), lang)
	if err != nil {
		log.FromContext(ctx).Error(err.Error())
		c.ResponseError(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	c.ResponseJson(w, r, city)
}

func (c *GeoIpController) getGeoNameFilter(r *http.Request) entity.GeoNameFilter {
	codes, _ := gost.GetQueryOptionSlice[string](r, "country-codes")
	return entity.GeoNameFilter{
		CountryCodes: codes,
	}
}

// @Summary country
// @Produce json
// @Tags geo IP
// @Param addr path string true "ip or hostname"
// @Success 200 {object} []entity.GeoNameCountry
// @Failure 400 {string} string "error"
// @Failure 500 {string} string "error"
// @Router /geoname/country [get]
func (c *GeoIpController) GetGeoNameCountriesHandler(w http.ResponseWriter, r *http.Request) {
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
// @Tags geo IP
// @Param addr path string true "ip or hostname"
// @Success 200 {object} []entity.AdminSubdivision
// @Failure 400 {string} string "error"
// @Failure 500 {string} string "error"
// @Router /geoname/subdivision [get]
func (c *GeoIpController) GetGeoNameSubdivisionsHandler(w http.ResponseWriter, r *http.Request) {
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
// @Tags geo IP
// @Param addr path string true "ip or hostname"
// @Success 200 {object} []entity.Geoname
// @Failure 400 {string} string "error"
// @Failure 500 {string} string "error"
// @Router /geoname/city [get]
func (c *GeoIpController) GetGeoNameCitiesHandler(w http.ResponseWriter, r *http.Request) {
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
