package microservice

import (
	"github.com/bldsoft/geos/pkg/config"
	"github.com/bldsoft/geos/pkg/controller"
	"github.com/bldsoft/geos/pkg/controller/rest"
	"github.com/bldsoft/geos/pkg/repository"
	"github.com/bldsoft/geos/pkg/service"
	gost "github.com/bldsoft/gost/controller"
	"github.com/bldsoft/gost/server"
	"github.com/go-chi/chi/v5"
)

type Microservice struct {
	config *config.Config

	geoService controller.GeoIpService

	grpcMicroservice *GrpcMicroservice
}

func NewMicroservice(config config.Config) *Microservice {
	srv := &Microservice{
		config: &config,
	}
	srv.initServices()
	return srv
}

func (m *Microservice) initServices() {
	rep := repository.NewGeoIpRepository(m.config.GeoDbPath)
	m.geoService = service.NewGeoService(rep)

	m.grpcMicroservice = NewGrpcMicroservice(m.config.GrpcAddr(), m.geoService)
}

func (m *Microservice) BuildRoutes(router chi.Router) {
	router.Route("/geoip/v2.1", func(r chi.Router) {
		r.Get("/env", gost.GetEnvHandler(m.config, nil))
		r.Get("/version", gost.GetVersionHandler)

		controller := rest.NewGeoIpController(m.geoService)
		r.Get("/country/{ip}", controller.GetCityHandler)
		r.Get("/city/{ip}", controller.GetCityHandler)
	})
}

func (m *Microservice) GetAsyncRunners() []server.AsyncRunner {
	return []server.AsyncRunner{m.grpcMicroservice}
}