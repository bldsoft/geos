package main

import (
	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/server"
)

type Config struct {
	Server server.Config
	Log    log.Config
}

// SetDefaults ...
func (c *Config) SetDefaults() {}

// Validate ...
func (c *Config) Validate() error {
	return nil
}
