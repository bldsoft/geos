package main

import (
	"github.com/bldsoft/gost/config"
	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/server"
	"github.com/bldsoft/gost/version"
)

func main() {
	var cfg Config
	config.ReadConfig(&cfg, "")
	log.InitLogger(&cfg.Log)

	log.Infof("Geos v%s", version.Version)
	log.Logf("ENVIRONMENT:\n***\n%s***", config.FormatEnv(&cfg))

	server.NewServer(cfg.Server, NewMicroservice(cfg)).
		UseDefaultMiddlewares().
		Start()
}
