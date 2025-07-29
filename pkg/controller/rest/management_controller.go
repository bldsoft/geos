package rest

import (
	"errors"
	"net/http"

	"github.com/bldsoft/geos/pkg/controller"
	"github.com/bldsoft/geos/pkg/service"
	"github.com/bldsoft/geos/pkg/utils"
	gost "github.com/bldsoft/gost/controller"
	"github.com/go-chi/chi/v5"
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
	db := chi.URLParam(r, "db")
	update, err := c.geoIpService.CheckUpdates(r.Context(), service.DBType(db))
	if err != nil {
		c.ResponseError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	c.ResponseJson(w, r, update)
}

func (c *ManagementController) UpdateGeoIPHandler(w http.ResponseWriter, r *http.Request) {
	db := chi.URLParam(r, "db")
	err := c.geoIpService.StartUpdate(r.Context(), service.DBType(db))
	if err != nil {
		if errors.Is(err, utils.ErrUpdateInProgress) {
			c.ResponseError(w, err.Error(), http.StatusConflict)
		} else {
			c.ResponseError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	c.ResponseOK(w)
}

func (c *ManagementController) CheckGeonamesUpdatesHandler(w http.ResponseWriter, r *http.Request) {
	updates, err := c.geoNameService.CheckUpdates(r.Context())
	if err != nil {
		c.ResponseError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	c.ResponseJson(w, r, updates)
}

func (c *ManagementController) UpdateGeonamesHandler(w http.ResponseWriter, r *http.Request) {
	err := c.geoNameService.StartUpdate(r.Context())
	if err != nil {
		if errors.Is(err, utils.ErrUpdateInProgress) {
			c.ResponseError(w, err.Error(), http.StatusConflict)
		} else {
			c.ResponseError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	c.ResponseOK(w)
}
