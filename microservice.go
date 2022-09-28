package main

import (
	"github.com/bldsoft/geos/controller/rest"
	"github.com/bldsoft/geos/repository"
	"github.com/bldsoft/geos/service"
	"github.com/bldsoft/gost/controller"
	"github.com/bldsoft/gost/server"
	"github.com/go-chi/chi/v5"
)

type Microservice struct {
	controller.BaseController
	config *Config

	geoService *service.GeoService
}

func NewMicroservice(config Config) *Microservice {
	srv := &Microservice{config: &config}
	srv.initServices()

	return srv
}

func (m *Microservice) initServices() {
	rep := repository.NewGeoRepository(m.config.GeoDbPath)
	m.geoService = service.NewGeoService(rep)
}

func (m *Microservice) BuildRoutes(router chi.Router) {

	router.Route("/geoip/v2.1", func(r chi.Router) {
		controller := rest.NewGeoGeoController(m.geoService)
		r.Get("/country/{ip}", controller.GetCityHandler)
		r.Get("/city/{ip}", controller.GetCityHandler)
	})
}

func (m *Microservice) GetAsyncRunners() []server.AsyncRunner {
	return nil
}
