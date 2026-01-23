package rest

import (
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/bldsoft/geos/pkg/controller"
	_ "github.com/bldsoft/geos/pkg/entity"
	"github.com/bldsoft/geos/pkg/repository"
	"github.com/bldsoft/geos/pkg/service"
	"github.com/bldsoft/geos/pkg/utils"
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
		c.responseError(w, r, err)
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
		c.responseError(w, r, err)
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
		c.responseError(w, r, err)
		return
	}
	c.ResponseJson(w, r, city)
}

// @Summary hosting
// @Produce json
// @Tags geo IP
// @Param addr path string true "ip or hostname"
// @Success 200 {object} entity.Hosting
// @Failure 400 {string} string "error"
// @Failure 500 {string} string "error"
// @Router /hosting/{addr} [get]
func (c *GeoIpController) GetHostingHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	hosting, err := c.geoIpService.Hosting(ctx, c.address(r))
	if err != nil {
		c.responseError(w, r, err)
		return
	}
	c.ResponseJson(w, r, hosting)
}

// @Summary geoip database dump
// @Security ApiKeyAuth
// @Deprecated
// @Produce text/csv
// @Tags geo IP
// @Success 200 {object} string
// @Failure 400 {string} string "error"
// @Failure 500 {string} string "error"
// @Failure 503 {string} string "error"
// @Router /dump [get]
func (c *GeoIpController) GetDumpHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	format := repository.DumpFormatCSV
	if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		format = repository.DumpFormatGzippedCSV
		w.Header().Set("Content-Encoding", "gzip")
	}

	db, err := c.geoIpService.Database(ctx, repository.MaxmindDBTypeCity, service.DumpFormat(format))
	if err != nil {
		c.responseError(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "text/csv")
	if _, err := io.Copy(w, db.Data); err != nil {
		log.FromContext(ctx).ErrorWithFields(log.Fields{"err": err}, "Failed to send csv database")
	}
}

// @Summary maxmind mmdb database
// @Security ApiKeyAuth
// @Produce octet-stream
// @Param db path string true "db type" Enums(city,isp)
// @Tags geo IP
// @Success 200 {object} string
// @Failure 400 {string} string "error"
// @Failure 500 {string} string "error"
// @Failure 503 {string} string "error"
// @securityDefinitions.apikey ApiKeyAuth
// @Router /dump/{db}/mmdb [get]
func (c *GeoIpController) GetMMDBDatabaseHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	db := chi.URLParam(r, "db")
	database, err := c.geoIpService.Database(ctx, service.DBType(db), repository.DumpFormatMMDB)
	if err != nil {
		c.responseError(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename="+database.FileName())
	if _, err := io.Copy(w, database.Data); err != nil {
		log.FromContext(ctx).ErrorWithFields(log.Fields{"err": err}, "Failed to send csv database")
	}
}

// @Summary maxmind csv database. It's generated from the mmdb file, so the result may differ from those that are officially supplied
// @Security ApiKeyAuth
// @Produce text/csv
// @Param db path string true "db type" Enums(city,isp)
// @Tags geo IP
// @Success 200 {object} string
// @Failure 400 {string} string "error"
// @Failure 500 {string} string "error"
// @Failure 503 {string} string "error"
// @securityDefinitions.apikey ApiKeyAuth
// @Router /dump/{db}/csv [get]
func (c *GeoIpController) GetCSVDatabaseHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	db := chi.URLParam(r, "db")
	format := repository.DumpFormatCSV
	if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		format = repository.DumpFormatGzippedCSV
		w.Header().Set("Content-Encoding", "gzip")
	}
	database, err := c.geoIpService.Database(ctx, service.DBType(db), format)
	if err != nil {
		c.responseError(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename="+database.FileName())
	if _, err := io.Copy(w, database.Data); err != nil {
		log.FromContext(ctx).ErrorWithFields(log.Fields{"err": err}, "Failed to send csv database")
	}
}

// @Summary maxmind database metadata
// @Security ApiKeyAuth
// @Produce json
// @Param db path string true "db type" Enums(city,isp)
// @Tags geo IP
// @Success 200 {object} entity.MetaData
// @Failure 400 {string} string "error"
// @Failure 500 {string} string "error"
// @Failure 503 {string} string "error"
// @Router /dump/{db}/metadata [get]
func (c *GeoIpController) GetDatabaseMetaHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	db := chi.URLParam(r, "db")
	metadata, err := c.geoIpService.MetaData(ctx, service.DBType(db))
	if err != nil {
		c.responseError(w, r, err)
		return
	}
	c.ResponseJson(w, r, metadata)
}

func (c *GeoIpController) responseError(w http.ResponseWriter, r *http.Request, err error) {
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
