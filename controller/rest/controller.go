package rest

import (
	"net"
	"net/http"
	"strings"

	"github.com/bldsoft/geos/controller"
	gost "github.com/bldsoft/gost/controller"
	"github.com/bldsoft/gost/log"
	"github.com/go-chi/chi/v5"
)

type GeoController struct {
	gost.BaseController
	service controller.GeoService
}

func NewGeoGeoController(service controller.GeoService) (c *GeoController) {
	return &GeoController{service: service}
}

func (c *GeoController) ip(r *http.Request) net.IP {
	ipStr := chi.URLParam(r, "ip")
	if ipStr == "me" {
		return net.ParseIP(strings.Split(r.RemoteAddr, ":")[0])
	}
	return net.ParseIP(ipStr)
}

func (c *GeoController) GetCityHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	city, err := c.service.City(ctx, c.ip(r))
	if err != nil {
		log.FromContext(ctx).Error(err.Error())
		c.ResponseError(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	c.ResponseJson(w, r, city)
}

func (c *GeoController) GetCountryHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	country, err := c.service.Country(ctx, c.ip(r))
	if err != nil {
		log.FromContext(ctx).Error(err.Error())
		c.ResponseError(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	c.ResponseJson(w, r, country)
}
