package microservice

import (
	"net/http"

	"github.com/bldsoft/geos/pkg/config"
	"github.com/bldsoft/geos/pkg/controller"
	"github.com/bldsoft/geos/pkg/controller/rest"
	"github.com/bldsoft/geos/pkg/repository"
	"github.com/bldsoft/geos/pkg/service"
	"github.com/bldsoft/geos/pkg/storage"
	"github.com/bldsoft/gost/auth"
	"github.com/bldsoft/gost/consul"
	gost "github.com/bldsoft/gost/controller"
	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/server"
	"github.com/go-chi/chi/v5"
)

const (
	BaseApiPath      = "/geoip"
	APIKey           = "GEOS-API-Key"
	ConsulAPIMetaKey = "api-key"
)

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

	geoNameStorage := storage.NewGeoNamesStorage(m.config.GeoNameDumpDirPath)
	geoNameRep := repository.NewGeoNamesRepository(geoNameStorage)
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
		discovery.SetMetadata(ConsulAPIMetaKey, m.config.ApiKey)
	}
}

func (m *Microservice) BuildRoutes(router chi.Router) {
	router.Route(BaseApiPath, func(r chi.Router) {
		r.Get("/ping", gost.GetPingHandler)
		r.Get("/env", gost.GetEnvHandler(m.config, nil))
		r.Get("/version", gost.GetVersionHandler)

		geoIpController := rest.NewGeoIpController(m.geoIpService)
		r.Get("/country/{addr}", geoIpController.GetCountryHandler)
		r.Get("/city/{addr}", geoIpController.GetCityHandler)
		r.Get("/city-lite/{addr}", geoIpController.GetCityLiteHandler)
		r.With(m.ApiKeyMiddleware()).Get("/dump", geoIpController.GetDumpHandler)

		geoNameController := rest.NewGeoNameController(m.geoNameService)
		r.Route("/geoname", func(r chi.Router) {
			r.Get("/continent", geoNameController.GetGeoNameContinentsHandler)
			r.Get("/country", geoNameController.GetGeoNameCountriesHandler)
			r.Get("/subdivision", geoNameController.GetGeoNameSubdivisionsHandler)
			r.Get("/city", geoNameController.GetGeoNameCitiesHandler)
			r.With(m.ApiKeyMiddleware()).Get("/dump", geoNameController.GetDumpHandler)
		})
	})
}

func (m *Microservice) ApiKeyMiddleware() func(next http.Handler) http.Handler {
	return auth.ApiKeyMiddleware(APIKey, m.config.ApiKey)
}

func (m *Microservice) GetAsyncRunners() []server.AsyncRunner {
	return m.asyncRunners
}
