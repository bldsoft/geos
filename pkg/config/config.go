package config

import (
	"fmt"
	"os"

	"github.com/bldsoft/gost/clickhouse"
	"github.com/bldsoft/gost/config"
	"github.com/bldsoft/gost/discovery/common"
	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/server"
)

const (
	ServiceName = "geos"
)

type Config struct {
	Server server.Config
	Log    log.Config

	Clickhouse clickhouse.Config `mapstructure:"CLICKHOUSE"`
	LogExport  clickhouse.LogExporterConfig

	GRPCServiceBindAddress config.Address `mapstructure:"GRPC_SERVICE_BIND_ADDRESS" description:"Service configuration related to what address bind to and port to listen"`
	GRPCServiceAddress     config.Address `mapstructure:"GRPC_SERVICE_ADDRESS" description:"GRPC public address"`

	GeoDbPath    string `mapstructure:"GEOIP_DB_PATH" description:"Path to GeoLite2 or GeoIP2 city database"`
	GeoDbISPPath string `mapstructure:"GEOIP_DB_ISP_PATH" description:"Path to GeoIP2 ISP database"`

	Discovery common.Config `mapstructure:"DISCOVERY"`

	GeoNameDumpDirPath  string `mapstructure:"GEONAME_DUMP_DIR" description:"The path to the directory where the GeoNames dumps are located (countryInfo.txt, admin1CodesASKII.txt, cities5000.zip). If variable isn't set, GeoNames api will be disabled. The dumps will be loaded when service starts, if something is missing"`
	GeoIPCsvDumpDirPath string `mapstructure:"GEOIP_DUMP_DIR" description:"The path to the directory where the csv ip database is located. If the variable is set and the csv file is missing, the service will generate it from the mmdb when it starts."`
	ApiKey              string `mapstructure:"API_KEY" description:"API key for dumps used for importing into other databases"`
}

func (c *Config) NeedGrpc() bool {
	return len(c.GRPCServiceBindAddress) > 0
}

// SetDefaults ...
func (c *Config) SetDefaults() {
	c.Server.ServiceName = ServiceName
	c.Server.ServiceBindHost = "0.0.0.0"
	c.Server.ServiceBindPort = 8505
	c.GRPCServiceBindAddress = "0.0.0.0:8506"
	c.GRPCServiceAddress = c.GRPCServiceBindAddress
	c.Log.Color = false
	c.GeoDbPath = "../../db.mmdb"
	c.ApiKey = "Dfga4pBfeRsMnxesWmY8eNBCW2Zf46kL"

	c.Clickhouse.Dsn = ""
	c.GeoNameDumpDirPath = "/data/geoname"
}

// Validate ...
func (c *Config) Validate() error {
	if _, err := os.Stat(c.GeoDbPath); err != nil {
		return fmt.Errorf("GEO_DB_PATH %s: %w", c.GeoDbPath, err)
	}

	return nil
}
