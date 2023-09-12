package config

import (
	"fmt"
	"os"

	"github.com/bldsoft/gost/discovery/common"
	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/server"
)

const (
	ServiceName     = "geos"
	GrpcServiceName = "grpc_" + ServiceName
	RestServiceName = "rest_" + ServiceName
)

type Config struct {
	Server       server.Config
	Log          log.Config
	GrpcPort     int    `mapstructure:"GRPC_SERVICE_PORT" description:"gRPC service port (0 - disabled)"`
	GeoDbPath    string `mapstructure:"GEOIP_DB_PATH" description:"Path to GeoLite2 or GeoIP2 city database"`
	GeoDbISPPath string `mapstructure:"GEOIP_DB_ISP_PATH" description:"Path to GeoIP2 ISP database"`

	GrpcDiscovery common.Config `mapstructure:"DISCOVERY_GRPC"`
	RestDiscovery common.Config `mapstructure:"DISCOVERY_REST"`

	GeoNameDumpDirPath  string `mapstructure:"GEONAME_DUMP_DIR" description:"The path to the directory where the GeoNames dumps are located (countryInfo.txt, admin1CodesASKII.txt, cities5000.zip). If variable isn't set, GeoNames api will be disabled. The dumps will be loaded when service starts, if something is missing"`
	GeoIPCsvDumpDirPath string `mapstructure:"GEOIP_DUMP_DIR" description:"The path to the directory where the csv ip database is located. If the variable is set and the csv file is missing, the service will generate it from the mmdb when it starts."`
	ApiKey              string `mapstructure:"API_KEY" description:"API key for dumps used for importing into other databases"`
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

	c.GrpcDiscovery.ServiceName = GrpcServiceName
	c.GrpcDiscovery.ServiceProto = ""
	c.RestDiscovery.ServiceName = RestServiceName

	c.ApiKey = "Dfga4pBfeRsMnxesWmY8eNBCW2Zf46kL"
}

// Validate ...
func (c *Config) Validate() error {
	if _, err := os.Stat(c.GeoDbPath); err != nil {
		return fmt.Errorf("GEO_DB_PATH %s: %w", c.GeoDbPath, err)
	}
	return nil
}
