package main

import (
	"fmt"

	"github.com/bldsoft/geos/controller"
	"github.com/bldsoft/geos/controller/grpc"
	"github.com/bldsoft/geos/controller/rest"
	"github.com/bldsoft/geos/repository"
	"github.com/bldsoft/geos/service"
	"github.com/bldsoft/gost/server"
	"github.com/go-chi/chi/v5"
)

type Microservice struct {
	config *Config

	geoService controller.GeoService

	grpcMicroservice *grpc.Server
}

func NewMicroservice(config Config) *Microservice {
	srv := &Microservice{
		config: &config,
	}
	srv.initServices()
	return srv
}

func (m *Microservice) initServices() {
	rep := repository.NewGeoRepository(m.config.GeoDbPath)
	m.geoService = service.NewGeoService(rep)

	grpcAddr := fmt.Sprintf("%s:%d", m.config.Server.Host, m.config.GrpcPort)
	m.grpcMicroservice = grpc.NewServer(grpcAddr, m.geoService)
}

func (m *Microservice) BuildRoutes(router chi.Router) {
	router.Route("/geoip/v2.1", func(r chi.Router) {
		controller := rest.NewGeoGeoController(m.geoService)
		r.Get("/country/{ip}", controller.GetCityHandler)
		r.Get("/city/{ip}", controller.GetCityHandler)
	})
}

func (m *Microservice) GetAsyncRunners() []server.AsyncRunner {
	return []server.AsyncRunner{m.grpcMicroservice}
}
