package config

import (
	"fmt"
	"os"
	"time"

	"github.com/bldsoft/gost/config"
	"github.com/bldsoft/gost/discovery/common"
	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/server"
	"github.com/bldsoft/gost/utils"
)

const (
	ServiceName = "geos"
)

type Config struct {
	Server server.Config `mapstructure:"REST"`
	Log    log.Config    `mapstructure:"REST"`

	GRPCServiceBindAddress config.Address `mapstructure:"GRPC_SERVICE_BIND_ADDRESS" description:"Service configuration related to what address bind to and port to listen"`
	GRPCServiceAddress     config.Address `mapstructure:"GRPC_SERVICE_ADDRESS" description:"GRPC public address"`

	GeoDbPath    string `mapstructure:"GEOIP_DB_PATH" description:"Path to GeoLite2 or GeoIP2 city database"`
	GeoDbISPPath string `mapstructure:"GEOIP_DB_ISP_PATH" description:"Path to GeoIP2 ISP database"`

	DeprecatedConsul DeprecatedConsulConfig // TODO: remove

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
	c.Server.ServiceBindAddress = "0.0.0.0:8505"
	c.GRPCServiceAddress = "0.0.0.0:8506"

	c.Log.Color = false
	c.GeoDbPath = "../../db.mmdb"

	c.ApiKey = "Dfga4pBfeRsMnxesWmY8eNBCW2Zf46kL"
}

// Validate ...
func (c *Config) Validate() error {
	if _, err := os.Stat(c.GeoDbPath); err != nil {
		return fmt.Errorf("GEO_DB_PATH %s: %w", c.GeoDbPath, err)
	}

	c.setFromDeprecated()

	return nil
}

type DeprecatedConsulConfig struct {
	ConsulAddr   string              `mapstructure:"CONSUL_ADDRESS" description:"DEPRECATED. Address of the Consul server"`
	ConsulScheme string              `mapstructure:"CONSUL_SCHEME" description:"DEPRECATED. URI scheme for the Consul server"`
	Token        config.HiddenString `mapstructure:"CONSUL_TOKEN" description:" DEPRECATED. Token is used to provide a per-request ACL token"`

	GrpcServiceID      string        `mapstructure:"CONSUL_GRPC_SERVICE_ID" description:"DEPRECATED. The ID of the service. If empty, a random one will be generated"`
	GrpcCluster        string        `mapstructure:"CONSUL_GRPC_CLUSTER" description:"DEPRECATED. The name of the service to register"`
	GrpcServiceAddr    string        `mapstructure:"CONSUL_GRPC_SERVICE_ADDRESS" description:"DEPRECATED. The address of the service. If it's empty the service doesn't register in consul"`
	GrpcServicePort    int           `mapstructure:"CONSUL_GRPC_SERVICE_PORT" description:"DEPRECATED. The port of the service"`
	GrpcHealthCheckTTL time.Duration `mapstructure:"CONSUL_GRPC_HEALTH_CHECK_TTL" description:"DEPRECATED. Check TTL"`
	GrpcDeregisterTTL  time.Duration `mapstructure:"CONSUL_GRPC_DEREREGISTER_TTL" description:"DEPRECATED. If a check is in the critical state for more than this configured value,	then the service will automatically be deregistered"`

	RestServiceID      string        `mapstructure:"CONSUL_REST_SERVICE_ID" description:"DEPRECATED. The ID of the service. If empty, a random one will be generated"`
	RestCluster        string        `mapstructure:"CONSUL_REST_CLUSTER" description:"DEPRECATED. The name of the service to register"`
	RestServiceAddr    string        `mapstructure:"CONSUL_REST_SERVICE_ADDRESS" description:"DEPRECATED. The address of the service. If it's empty the service doesn't register in consul"`
	RestServicePort    int           `mapstructure:"CONSUL_REST_SERVICE_PORT" description:"DEPRECATED. The port of the service"`
	RestHealthCheckTTL time.Duration `mapstructure:"CONSUL_REST_HEALTH_CHECK_TTL" description:"DEPRECATED. Check TTL"`
	RestDeregisterTTL  time.Duration `mapstructure:"CONSUL_REST_DEREREGISTER_TTL" description:"DEPRECATED. If a check is in the critical state for more than this configured value,	then the service will automatically be deregistered"`
}

func (c *DeprecatedConsulConfig) SetDefaults() {
	c.ConsulScheme = "http"

	c.GrpcCluster = ServiceName
	c.GrpcHealthCheckTTL = 30 * time.Second
	c.GrpcDeregisterTTL = 30 * time.Second
}

func (c *DeprecatedConsulConfig) Validate() error {
	if len(c.GrpcServiceID) == 0 {
		c.GrpcServiceID = utils.RandString(32)
	}
	if len(c.RestServiceID) == 0 {
		c.RestServiceID = utils.RandString(32)
	}
	return nil
}

func (c *Config) setFromDeprecated() {
	if c.DeprecatedConsul.RestServiceAddr != "" {
		c.Discovery.DiscoveryType = common.DiscoveryTypeConsul

		c.Discovery.Consul.ConsulAddr = config.HttpAddress(fmt.Sprintf("%s://%s", c.DeprecatedConsul.ConsulScheme, c.DeprecatedConsul.ConsulAddr))
		c.Discovery.Consul.Token = c.DeprecatedConsul.Token

		c.Discovery.Consul.HealthCheckTTL = c.DeprecatedConsul.GrpcHealthCheckTTL
	}

}
