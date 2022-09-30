package rest

import (
	"net"
	"net/http"
	"strings"

	"github.com/bldsoft/geos/pkg/controller"
	gost "github.com/bldsoft/gost/controller"
	"github.com/bldsoft/gost/log"
	"github.com/go-chi/chi/v5"
)

type GeoIpController struct {
	gost.BaseController
	service controller.GeoIpService
}

func NewGeoIpController(service controller.GeoIpService) (c *GeoIpController) {
	return &GeoIpController{service: service}
}

func (c *GeoIpController) ip(r *http.Request) net.IP {
	ipStr := chi.URLParam(r, "ip")
	if ipStr == "me" {
		return net.ParseIP(strings.Split(r.RemoteAddr, ":")[0])
	}
	return net.ParseIP(ipStr)
}

func (c *GeoIpController) GetCityHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	city, err := c.service.City(ctx, c.ip(r))
	if err != nil {
		log.FromContext(ctx).Error(err.Error())
		c.ResponseError(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	c.ResponseJson(w, r, city)
}

func (c *GeoIpController) GetCountryHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	country, err := c.service.Country(ctx, c.ip(r))
	if err != nil {
		log.FromContext(ctx).Error(err.Error())
		c.ResponseError(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	c.ResponseJson(w, r, country)
}