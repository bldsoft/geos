package rest

import (
	"errors"
	"net/http"

	"github.com/bldsoft/geos/pkg/controller"
	"github.com/bldsoft/geos/pkg/storage/state"
	"github.com/bldsoft/geos/pkg/utils"
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
		return
	}
	c.ResponseJson(w, r, exist)
}

func (c *ManagementController) UpdateGeoIPHandler(w http.ResponseWriter, r *http.Request) {
	updates, err := c.geoIpService.Download(r.Context())
	if err != nil {
		if errors.Is(err, utils.ErrUpdateInProgress) {
			c.ResponseError(w, err.Error(), http.StatusConflict)
		} else {
			c.ResponseError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	c.ResponseJson(w, r, updates)
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
	updates, err := c.geoNameService.Download(r.Context())
	if err != nil {
		if errors.Is(err, utils.ErrUpdateInProgress) {
			c.ResponseError(w, err.Error(), http.StatusConflict)
		} else {
			c.ResponseError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	c.ResponseJson(w, r, updates)
}

func (c *ManagementController) GetGeosStateHandler(w http.ResponseWriter, r *http.Request) {
	result := &state.GeosState{}

	geoIpState := c.geoIpService.State()
	if geoIpState != nil {
		result.Add(geoIpState)
	}

	geoNameState := c.geoNameService.State()
	if geoNameState != nil {
		result.Add(geoNameState)
	}

	c.ResponseJson(w, r, result)
}
