package main

import (
	"github.com/bldsoft/geos/pkg/config"
	"github.com/bldsoft/geos/pkg/microservice"
	gost "github.com/bldsoft/gost/config"
	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/server"
	"github.com/bldsoft/gost/version"
)

//go:generate swag init -d ../.. --parseDependency -g cmd/geos/main.go -o ../../api/rest

// @title Geos API
// @version 2.1
// @description Content delivery service
// @schemes http
// @license.name ᅠ
// @BasePath /geoip/v2.1

func main() {
	var cfg config.Config
	gost.ReadConfig(&cfg, "")
	log.InitLogger(&cfg.Log)

	log.Infof("Geos v%s", version.Version)
	log.Logf("ENVIRONMENT:\n***\n%s***", gost.FormatEnv(&cfg))

	server.NewServer(cfg.Server, microservice.NewMicroservice(cfg)).
		UseDefaultMiddlewares().
		Start()
}
