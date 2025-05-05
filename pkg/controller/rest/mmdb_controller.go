package rest

import (
	"net/http"

	"github.com/bldsoft/geos/pkg/controller"
	gost "github.com/bldsoft/gost/controller"
	"github.com/bldsoft/gost/log"
)

type MMDBController struct {
	gost.BaseController
	mmdbService controller.MMDBService
}

func NewMMDBController(mmdbServic controller.MMDBService) *MMDBController {
	return &MMDBController{mmdbService: mmdbServic}
}

func (c *MMDBController) CheckUpdateHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var (
		cityDbUpdate, ispDbUpdate bool
		err                       error
	)

	if cityDbUpdate, err = c.mmdbService.CheckCityDbUpdates(ctx); err != nil {
		log.FromContext(ctx).ErrorWithFields(log.Fields{"error": err}, "Failed to check city database updates")
		c.ResponseError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if ispDbUpdate, err = c.mmdbService.CheckISPDbUpdates(ctx); err != nil {
		log.FromContext(ctx).ErrorWithFields(log.Fields{"error": err}, "Failed to check ISP database updates")
		c.ResponseError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	c.ResponseJson(w, r, map[string]bool{
		"city database updates": cityDbUpdate,
		"isp database updates":  ispDbUpdate,
	}, false)
}

func (c *MMDBController) UpdateHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := c.mmdbService.DownloadCityDb(ctx, true); err != nil {
		log.FromContext(ctx).ErrorWithFields(log.Fields{"error": err}, "Failed to download city database")
		c.ResponseError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := c.mmdbService.DownloadISPDb(ctx, true); err != nil {
		log.FromContext(ctx).ErrorWithFields(log.Fields{"error": err}, "Failed to download ISP database")
		c.ResponseError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	c.ResponseJson(w, r, "OK", false)
}
