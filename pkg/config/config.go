package config

import (
	"fmt"
	"os"

	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/server"
)

type Config struct {
	Server    server.Config
	Log       log.Config
	GrpcPort  int    `mapstructure:"GRPC SERVICE_PORT" description:"gRPC service port"`
	GeoDbPath string `mapstructure:"GEO_DB_PATH" description:"Path to GeoLite2 or GeoIP2 databases"`
}

// SetDefaults ...
func (c *Config) SetDefaults() {
	c.GeoDbPath = "../../GeoLite2-City.mmdb"
	c.GrpcPort = 3001
}

// Validate ...
func (c *Config) Validate() error {
	if _, err := os.Stat(c.GeoDbPath); err != nil {
		return fmt.Errorf("GEO_DB_PATH %s: %w", c.GeoDbPath, err)
	}
	return nil
}
