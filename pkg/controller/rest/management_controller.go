package rest

import (
	"net/http"

	"github.com/bldsoft/geos/pkg/controller"
	gost "github.com/bldsoft/gost/controller"
)

type ManagementController struct {
	gost.BaseController
	geoIpService   controller.GeoIpService
	GeoNameService controller.GeoNameService
}

func NewManagementController(geoIpService controller.GeoIpService, geoNameService controller.GeoNameService) *ManagementController {
	return &ManagementController{geoIpService: geoIpService, GeoNameService: geoNameService}
}

func (c *ManagementController) CheckUpdatesHandler(w http.ResponseWriter, r *http.Request) {
	exist, err := c.geoIpService.CheckUpdates(r.Context())
	if err != nil {
		c.ResponseError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	c.ResponseJson(w, r, exist)
}

func (c *ManagementController) UpdateHandler(w http.ResponseWriter, r *http.Request) {
	if err := c.geoIpService.Download(r.Context(), true); err != nil {
		c.ResponseError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	c.ResponseJson(w, r, true)
}
