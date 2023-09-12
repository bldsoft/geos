|**Environment variable**|**Value**|**Description**|
|------------------------|---------|---------------|
|SERVICE_NAME|default_name|Unique service instance name<br/>The name is used to identify the service in logs|
|SERVICE_HOST|0.0.0.0|IP address, or a host name that can be resolved to IP addresses|
|SERVICE_PORT|8505|Service port|
|LOG_LEVEL|info|Log level|
|LOG_COLOR_ENABLED|false|Enable the colorized output|
|GRPC_SERVICE_PORT|0|gRPC service port (0 - disabled)|
|GEOIP_DB_PATH|../../db.mmdb|Path to GeoLite2 or GeoIP2 city database|
|GEOIP_DB_ISP_PATH||Path to GeoIP2 ISP database|
|CONSUL_ADDRESS||DEPRECATED. Address of the Consul server|
|CONSUL_SCHEME|http|DEPRECATED. URI scheme for the Consul server|
|CONSUL_TOKEN|| DEPRECATED. Token is used to provide a per-request ACL token|
|CONSUL_GRPC_SERVICE_ID||DEPRECATED. The ID of the service. If empty, a random one will be generated|
|CONSUL_GRPC_CLUSTER|grpc_geos|DEPRECATED. The name of the service to register|
|CONSUL_GRPC_SERVICE_ADDRESS||DEPRECATED. The address of the service. If it's empty the service doesn't register in consul|
|CONSUL_GRPC_SERVICE_PORT|0|DEPRECATED. The port of the service|
|CONSUL_GRPC_HEALTH_CHECK_TTL|30s|DEPRECATED. Check TTL|
|CONSUL_GRPC_DEREREGISTER_TTL|30s|DEPRECATED. If a check is in the critical state for more than this configured value,	then the service will automatically be deregistered|
|CONSUL_REST_SERVICE_ID||DEPRECATED. The ID of the service. If empty, a random one will be generated|
|CONSUL_REST_CLUSTER|rest_geos|DEPRECATED. The name of the service to register|
|CONSUL_REST_SERVICE_ADDRESS||DEPRECATED. The address of the service. If it's empty the service doesn't register in consul|
|CONSUL_REST_SERVICE_PORT|0|DEPRECATED. The port of the service|
|CONSUL_REST_HEALTH_CHECK_TTL|30s|DEPRECATED. Check TTL|
|CONSUL_REST_DEREREGISTER_TTL|30s|DEPRECATED. If a check is in the critical state for more than this configured value,	then the service will automatically be deregistered|
|DISCOVERY_GRPC_SERVICE_ID||The ID of the service. This must be unique in the cluster. If empty, a random one will be generated|
|DISCOVERY_GRPC_SERVICE_NAME|grpc_geos|The name of the service to register|
|DISCOVERY_GRPC_SERVICE_PROTO||The proto of the service|
|DISCOVERY_GRPC_SERVICE_HOST||The address of the service. If it's empty the service doesn't register in discovery|
|DISCOVERY_GRPC_SERVICE_PORT||The port of the service|
|DISCOVERY_GRPC_TYPE|none|Discovery type|
|DISCOVERY_GRPC_MEMBERLIST_HOST||Memberlist host|
|DISCOVERY_GRPC_MEMBERLIST_PORT|0|Meberlist port|
|DISCOVERY_GRPC_MEMBERLIST_CLUSTER_MEMBERS||Any existing member of the cluster to join it|
|DISCOVERY_GRPC_CONSUL_ADDRESS|127.0.0.1:8500|Address of the Consul server|
|DISCOVERY_GRPC_CONSUL_SCHEME|http|URI scheme for the Consul server|
|DISCOVERY_GRPC_CONSUL_TOKEN|| Token is used to provide a per-request ACL token|
|DISCOVERY_GRPC_CONSUL_HEALTH_CHECK_TTL|30s|Check TTL|
|DISCOVERY_GRPC_CONSUL_DEREREGISTER_TTL|30s|If a check is in the critical state for more than this configured value,	then the service will automatically be deregistered|
|DISCOVERY_REST_SERVICE_ID||The ID of the service. This must be unique in the cluster. If empty, a random one will be generated|
|DISCOVERY_REST_SERVICE_NAME|rest_geos|The name of the service to register|
|DISCOVERY_REST_SERVICE_PROTO|http|The proto of the service|
|DISCOVERY_REST_SERVICE_HOST||The address of the service. If it's empty the service doesn't register in discovery|
|DISCOVERY_REST_SERVICE_PORT||The port of the service|
|DISCOVERY_REST_TYPE|none|Discovery type|
|DISCOVERY_REST_MEMBERLIST_HOST||Memberlist host|
|DISCOVERY_REST_MEMBERLIST_PORT|0|Meberlist port|
|DISCOVERY_REST_MEMBERLIST_CLUSTER_MEMBERS||Any existing member of the cluster to join it|
|DISCOVERY_REST_CONSUL_ADDRESS|127.0.0.1:8500|Address of the Consul server|
|DISCOVERY_REST_CONSUL_SCHEME|http|URI scheme for the Consul server|
|DISCOVERY_REST_CONSUL_TOKEN|| Token is used to provide a per-request ACL token|
|DISCOVERY_REST_CONSUL_HEALTH_CHECK_TTL|30s|Check TTL|
|DISCOVERY_REST_CONSUL_DEREREGISTER_TTL|30s|If a check is in the critical state for more than this configured value,	then the service will automatically be deregistered|
|GEONAME_DUMP_DIR||The path to the directory where the GeoNames dumps are located (countryInfo.txt, admin1CodesASKII.txt, cities5000.zip). If variable isn't set, GeoNames api will be disabled. The dumps will be loaded when service starts, if something is missing|
|GEOIP_DUMP_DIR||The path to the directory where the csv ip database is located. If the variable is set and the csv file is missing, the service will generate it from the mmdb when it starts.|
|API_KEY|Dfga4pBfeRsMnxesWmY8eNBCW2Zf46kL|API key for dumps used for importing into other databases|
