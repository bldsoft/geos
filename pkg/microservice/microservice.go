package microservice

import (
	"github.com/bldsoft/geos/pkg/config"
	"github.com/bldsoft/geos/pkg/controller"
	"github.com/bldsoft/geos/pkg/controller/rest"
	"github.com/bldsoft/geos/pkg/repository"
	"github.com/bldsoft/geos/pkg/service"
	"github.com/bldsoft/gost/consul"
	gost "github.com/bldsoft/gost/controller"
	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/server"
	"github.com/go-chi/chi/v5"
)

const BaseApiPath = "/geoip"

type Microservice struct {
	config *config.Config

	geoIpService   controller.GeoIpService
	geoNameService controller.GeoNameService

	asyncRunners []server.AsyncRunner
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
	m.geoIpService = service.NewGeoIpService(rep)

	geoNameRep := repository.NewGeoNamesRepository()
	m.geoNameService = service.NewGeoNameService(geoNameRep)

	if m.config.NeedGrpc() {
		grpcService := NewGrpcMicroservice(m.config.GrpcAddr(), m.geoIpService, m.geoNameService)
		m.asyncRunners = append(m.asyncRunners, grpcService)

		if m.config.ConsulEnabled() {
			discovery := consul.NewDiscovery(m.config.GrpcConsulConfig())
			m.asyncRunners = append(m.asyncRunners, discovery)
		}
	} else {
		log.Info("gRPC is off")
	}

	if m.config.ConsulEnabled() {
		discovery := consul.NewDiscovery(m.config.RestConsulConfig())
		m.asyncRunners = append(m.asyncRunners, discovery)
	}
}

func (m *Microservice) BuildRoutes(router chi.Router) {
	router.Route(BaseApiPath, func(r chi.Router) {
		r.Get("/ping", gost.GetPingHandler)
		r.Get("/env", gost.GetEnvHandler(m.config, nil))
		r.Get("/version", gost.GetVersionHandler)

		controller := rest.NewGeoIpController(m.geoIpService, m.geoNameService)
		r.Get("/country/{addr}", controller.GetCountryHandler)
		r.Get("/city/{addr}", controller.GetCityHandler)

		r.Get("/city-lite/{addr}", controller.GetCityLiteHandler)

		r.Route("/geoname", func(r chi.Router) {
			r.Get("/country", controller.GetGeoNameCountriesHandler)
			r.Get("/subdivision", controller.GetGeoNameSubdivisionsHandler)
			r.Get("/city", controller.GetGeoNameCitiesHandler)
		})
	})
}

func (m *Microservice) GetAsyncRunners() []server.AsyncRunner {
	return m.asyncRunners
}
