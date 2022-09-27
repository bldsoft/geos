package main

import (
	"github.com/bldsoft/gost/controller"
	"github.com/bldsoft/gost/server"
	"github.com/go-chi/chi/v5"
)

type Microservice struct {
	controller.BaseController
	config *Config
}

func NewMicroservice(config Config) *Microservice {
	srv := &Microservice{config: &config}
	srv.initServices()

	return srv
}

func (m *Microservice) initServices() {
}

func (m *Microservice) BuildRoutes(router chi.Router) {
}

func (m *Microservice) GetAsyncRunners() []server.AsyncRunner {
	return nil
}
