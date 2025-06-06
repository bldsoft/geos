package rest

import (
	"net/http"

	"github.com/bldsoft/geos/pkg/controller"
	gost "github.com/bldsoft/gost/controller"
)

type ManagementController struct {
	gost.BaseController
	geoIpService   controller.GeoIpService
	geoNameService controller.GeoNameService
}

func NewManagementController(geoIpService controller.GeoIpService, geoNameService controller.GeoNameService) *ManagementController {
	return &ManagementController{geoIpService: geoIpService, geoNameService: geoNameService}
}

func (c *ManagementController) CheckGeoIPUpdatesHandler(w http.ResponseWriter, r *http.Request) {
	exist, err := c.geoIpService.CheckUpdates(r.Context())
	if err != nil {
		c.ResponseError(w, err.Error(), http.StatusInternalServerError)
	} else {
		c.ResponseJson(w, r, exist)
	}
}

func (c *ManagementController) UpdateGeoIPHandler(w http.ResponseWriter, r *http.Request) {
	if updates, err := c.geoIpService.Download(r.Context(), true); err != nil {
		c.ResponseError(w, err.Error(), http.StatusInternalServerError)
	} else {
		c.ResponseJson(w, r, updates)
	}
}

func (c *ManagementController) CheckGeonamesUpdatesHandler(w http.ResponseWriter, r *http.Request) {
	if updates, err := c.geoNameService.CheckUpdates(r.Context()); err != nil {
		c.ResponseError(w, err.Error(), http.StatusInternalServerError)
	} else {
		c.ResponseJson(w, r, updates)
	}
}

func (c *ManagementController) UpdateGeonamesHandler(w http.ResponseWriter, r *http.Request) {
	if updates, err := c.geoNameService.Download(r.Context()); err != nil {
		c.ResponseError(w, err.Error(), http.StatusInternalServerError)
	} else {
		c.ResponseJson(w, r, updates)
	}
}
