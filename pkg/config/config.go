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

	GeoDbSource           string `mapstructure:"GEOIP_DB_SOURCE" description:"Source to download GeoLite2 or GeoIP2 city database from"`
	GeoDbPatchesSource    string `mapstructure:"GEOIP_DB_PATCHES_SOURCE" description:"Source for downloading patches for city database (in .tar.gz)"`
	GeoDbISPSource        string `mapstructure:"GEOIP_DB_ISP_SOURCE" description:"Source to download GeoIP2 ISP database from"`
	GeoDbISPPatchesSource string `mapstructure:"GEOIP_DB_ISP_PATCHES_SOURCE" description:"Source for downloading custom ISP database patches (in .tar.gz)"`
	AutoUpdatePeriodSec   int    `mapstructure:"AUTO_UPDATE_PERIOD_SEC" description:"Amount of seconds to wait before trying to automatically update from the source"`

	GeoDbPath    string `mapstructure:"GEOIP_DB_PATH" description:"Path to GeoLite2 or GeoIP2 city database"`
	GeoDbISPPath string `mapstructure:"GEOIP_DB_ISP_PATH" description:"Path to GeoIP2 ISP database"`

	Discovery common.Config `mapstructure:"DISCOVERY"`

	GeoNameDumpDirPath   string `mapstructure:"GEONAME_DUMP_DIR" description:"The path to the directory where the GeoNames dumps are located (countryInfo.txt, admin1CodesASKII.txt, cities5000.zip). If variable isn't set, GeoNames api will be disabled. The dumps will be loaded when service starts, if something is missing"`
	GeoNamePatchesSource string `mapstructure:"GEONAME_PATCHES_SOURCE" description:"Source for downloading custom GeoNames patches (in .tar.gz)"`
	GeoIPCsvDumpDirPath  string `mapstructure:"GEOIP_DUMP_DIR" description:"The path to the directory where the csv ip database is located. If the variable is set and the csv file is missing, the service will generate it from the mmdb when it starts."`
	ApiKey               string `mapstructure:"API_KEY" description:"API key for dumps used for importing into other databases"`
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
}

// Validate ...
func (c *Config) Validate() error {
	if _, err := os.Stat(c.GeoDbPath); err != nil && len(c.GeoDbSource) == 0 {
		return fmt.Errorf("GEOIP_DB_PATH %s: %w", c.GeoDbPath, err)
	}

	if _, err := os.Stat(c.GeoDbISPPath); err != nil && len(c.GeoDbISPSource) == 0 {
		return fmt.Errorf("GEOIP_DB_ISP_PATH %s: %w", c.GeoDbISPPath, err)
	}

	return nil
}
