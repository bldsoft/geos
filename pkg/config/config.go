package config

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/bldsoft/gost/config"
	"github.com/bldsoft/gost/discovery/common"
	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/server"
	"github.com/bldsoft/gost/utils"
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

	DeprecatedConsul DeprecatedConsulConfig // TODO: remove

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
	c.RestDiscovery.ServiceName = RestServiceName

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

	c.GrpcCluster = GrpcServiceName
	c.GrpcHealthCheckTTL = 30 * time.Second
	c.GrpcDeregisterTTL = 30 * time.Second

	c.RestCluster = RestServiceName
	c.RestHealthCheckTTL = 30 * time.Second
	c.RestDeregisterTTL = 30 * time.Second
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
	if c.DeprecatedConsul.GrpcServiceAddr != "" {
		c.GrpcDiscovery.DiscoveryType = common.DiscoveryTypeConsul

		c.GrpcDiscovery.Consul.ConsulAddr = config.HttpAddress(fmt.Sprintf("%s://%s", c.DeprecatedConsul.ConsulScheme, c.DeprecatedConsul.ConsulAddr))
		c.GrpcDiscovery.Consul.Token = c.DeprecatedConsul.Token

		c.GrpcDiscovery.ServiceID = c.DeprecatedConsul.GrpcServiceID
		c.GrpcDiscovery.ServiceName = c.DeprecatedConsul.GrpcCluster
		if c.DeprecatedConsul.GrpcServicePort != 0 {
			c.GrpcDiscovery.ServiceAddr = config.Address(net.JoinHostPort(c.DeprecatedConsul.GrpcServiceAddr, strconv.Itoa(c.DeprecatedConsul.GrpcServicePort)))
		} else {
			c.GrpcDiscovery.ServiceAddr = config.Address(c.DeprecatedConsul.GrpcServiceAddr)
		}

		c.GrpcDiscovery.Consul.HealthCheckTTL = c.DeprecatedConsul.GrpcHealthCheckTTL
		c.GrpcDiscovery.Consul.DeregisterTTL = c.DeprecatedConsul.GrpcDeregisterTTL
	}

	if c.DeprecatedConsul.RestServiceAddr != "" {
		c.RestDiscovery.DiscoveryType = common.DiscoveryTypeConsul

		c.RestDiscovery.Consul.ConsulAddr = config.HttpAddress(fmt.Sprintf("%s://%s", c.DeprecatedConsul.ConsulScheme, c.DeprecatedConsul.ConsulAddr))

		c.RestDiscovery.Consul.Token = c.DeprecatedConsul.Token

		c.RestDiscovery.ServiceID = c.DeprecatedConsul.RestServiceID
		c.RestDiscovery.ServiceName = c.DeprecatedConsul.RestCluster
		if c.DeprecatedConsul.RestServicePort != 0 {
			c.RestDiscovery.ServiceAddr = config.Address(net.JoinHostPort(c.DeprecatedConsul.RestServiceAddr, strconv.Itoa(c.DeprecatedConsul.RestServicePort)))
		} else {
			c.RestDiscovery.ServiceAddr = config.Address(c.DeprecatedConsul.RestServiceAddr)
		}

		c.RestDiscovery.Consul.HealthCheckTTL = c.DeprecatedConsul.RestHealthCheckTTL
		c.RestDiscovery.Consul.DeregisterTTL = c.DeprecatedConsul.RestDeregisterTTL
	}

}
