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
	GeoDbPath string `mapstructure:"GEOIP_DB_PATH" description:"Path to GeoLite2 or GeoIP2 databases"`
}

func (c *Config) GrpcAddr() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.GrpcPort)
}

// SetDefaults ...
func (c *Config) SetDefaults() {
	c.GeoDbPath = "../../db.mmdb"
	c.GrpcPort = 3001
}

// Validate ...
func (c *Config) Validate() error {
	if _, err := os.Stat(c.GeoDbPath); err != nil {
		return fmt.Errorf("GEO_DB_PATH %s: %w", c.GeoDbPath, err)
	}
	return nil
}
