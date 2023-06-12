package config

import (
	"fmt"
	"os"

	"github.com/bldsoft/gost/consul"
	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/server"
)

const (
	ServiceName           = "geos"
	ConsulGrpcClusterName = "grpc_" + ServiceName
	ConsulRestClusterName = "rest_" + ServiceName
)

type Config struct {
	Server    server.Config
	Log       log.Config
	GrpcPort  int    `mapstructure:"GRPC_SERVICE_PORT" description:"gRPC service port (0 - disabled)"`
	GeoDbPath string `mapstructure:"GEOIP_DB_PATH" description:"Path to GeoLite2 or GeoIP2 databases"`

	Consul     consul.ConsulConfig  `mapstructure:"CONSUL"`
	GrpcConsul consul.ServiceConfig `mapstructure:"CONSUL_GRPC"`
	RestConsul consul.ServiceConfig `mapstructure:"CONSUL_REST"`

	GeoNameDumpDirPath string `mapstructure:"GEONAME_DUMP_DIR" description:"The path to the directory where the GeoNames dumps are located (countryInfo.txt, admin1CodesASKII.txt, cities5000.zip). If variable isn't set, GeoNames api will be disabled. The dumps will be loaded when service starts, if something is missing"`
	ApiKey             string `mapstructure:"API_KEY" description:"API key used to protect dumps that are used for importing into other databases"`
}

func (c *Config) ConsulEnabled() bool {
	return c.Consul.ConsulAddr != ""
}

func (c *Config) GrpcConsulConfig() consul.Config {
	return consul.Config{
		ConsulConfig:  c.Consul,
		ServiceConfig: c.GrpcConsul,
	}
}

func (c *Config) RestConsulConfig() consul.Config {
	return consul.Config{
		ConsulConfig:  c.Consul,
		ServiceConfig: c.RestConsul,
	}
}

func (c *Config) NeedGrpc() bool {
	return c.GrpcPort != 0
}

func (c *Config) GrpcAddr() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.GrpcPort)
}

// SetDefaults ...
func (c *Config) SetDefaults() {
	c.Server.Port = 8505
	c.Log.Color = false
	c.GeoDbPath = "../../db.mmdb"

	c.Consul.ConsulAddr = ""
	c.GrpcConsul.Cluster = ConsulGrpcClusterName
	c.RestConsul.Cluster = ConsulRestClusterName
	c.ApiKey = "Dfga4pBfeRsMnxesWmY8eNBCW2Zf46kL"
}

// Validate ...
func (c *Config) Validate() error {
	if _, err := os.Stat(c.GeoDbPath); err != nil {
		return fmt.Errorf("GEO_DB_PATH %s: %w", c.GeoDbPath, err)
	}
	return nil
}
