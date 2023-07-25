package rest

import (
	"net/http"

	"github.com/bldsoft/geos/pkg/controller"
	_ "github.com/bldsoft/geos/pkg/entity"
	"github.com/bldsoft/geos/pkg/repository"
	"github.com/bldsoft/geos/pkg/service"
	gost "github.com/bldsoft/gost/controller"
	"github.com/bldsoft/gost/log"
	"github.com/go-chi/chi/v5"
)

type GeoIpController struct {
	gost.BaseController
	geoIpService controller.GeoIpService
}

func NewGeoIpController(geoIpService controller.GeoIpService) (c *GeoIpController) {
	return &GeoIpController{geoIpService: geoIpService}
}

func (c *GeoIpController) address(r *http.Request) string {
	return chi.URLParam(r, "addr")
}

// @Summary city
// @Produce json
// @Tags geo IP
// @Param addr path string true "ip or hostname"
// @Param isp query bool false "include ISP info"
// @Success 200 {object} entity.City
// @Failure 400 {string} string "error"
// @Failure 500 {string} string "error"
// @Router /city/{addr} [get]
func (c *GeoIpController) GetCityHandler(w http.ResponseWriter, r *http.Request) {
	includeISP, _ := gost.GetQueryOption(r, "isp", false)
	ctx := r.Context()
	city, err := c.geoIpService.City(ctx, c.address(r), includeISP)
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

// @Summary geoip database dump
// @Security ApiKeyAuth
// @Deprecated
// @Produce text/csv
// @Tags geo IP
// @Param format query string false "format" "csvWithNames"
// @Success 200 {object} string
// @Failure 400 {string} string "error"
// @Failure 500 {string} string "error"
// @Router /dump [get]
func (c *GeoIpController) GetDumpHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	format, _ := gost.GetQueryOption(r, "format", repository.DumpFormatCSVWithNames)
	db, err := c.geoIpService.Database(ctx, repository.MaxmindDBTypeCity, service.DumpFormat(format))
	if err != nil {
		log.FromContext(ctx).Error(err.Error())
		c.ResponseError(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "text/csv")
	w.Write(db.Data)
}

// @Summary maxmind mmdb database
// @Security ApiKeyAuth
// @Produce octet-stream
// @Param db path string true "db type" Enums(city,isp)
// @Tags geo IP
// @Success 200 {object} string
// @Failure 400 {string} string "error"
// @Failure 500 {string} string "error"
// @securityDefinitions.apikey ApiKeyAuth
// @Router /dump/{db}/mmdb [get]
func (c *GeoIpController) GetMMDBDatabaseHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	db := chi.URLParam(r, "db")
	database, err := c.geoIpService.Database(ctx, service.DBType(db), repository.DumpFormatMMDB)
	if err != nil {
		log.FromContext(ctx).Error(err.Error())
		c.ResponseError(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename="+database.FileName())
	w.Write(database.Data)
}

// @Summary maxmind csv database. It's generated from the mmdb file, so the result may differ from those that are officially supplied
// @Security ApiKeyAuth
// @Produce text/csv
// @Param db path string true "db type" Enums(city,isp)
// @Param header query string false "include header" bool
// @Tags geo IP
// @Success 200 {object} string
// @Failure 400 {string} string "error"
// @Failure 500 {string} string "error"
// @securityDefinitions.apikey ApiKeyAuth
// @Router /dump/{db}/csv [get]
func (c *GeoIpController) GetCSVDatabaseHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	db := chi.URLParam(r, "db")
	includeHeader, _ := gost.GetQueryOption(r, "header", true)
	format := repository.DumpFormatCSVWithNames
	if !includeHeader {
		format = repository.DumpFormatCSV
	}
	database, err := c.geoIpService.Database(ctx, service.DBType(db), format)
	if err != nil {
		log.FromContext(ctx).Error(err.Error())
		c.ResponseError(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename="+database.FileName())
	w.Write(database.Data)
}

// @Summary maxmind database metadata
// @Security ApiKeyAuth
// @Produce json
// @Param db path string true "db type" Enums(city,isp)
// @Tags geo IP
// @Success 200 {object} entity.MetaData
// @Failure 400 {string} string "error"
// @Failure 500 {string} string "error"
// @Router /dump/{db}/metadata [get]
func (c *GeoIpController) GetDatabaseMetaHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	db := chi.URLParam(r, "db")
	database, err := c.geoIpService.Database(ctx, service.DBType(db), repository.DumpFormatMMDB)
	if err != nil {
		log.FromContext(ctx).Error(err.Error())
		c.ResponseError(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	c.ResponseJson(w, r, database.MetaData)
}
