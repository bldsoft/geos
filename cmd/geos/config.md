|**Environment variable**|**Value**|**Description**|
|------------------------|---------|---------------|
|SERVICE_NAME|default_name|Unique service instance name<br/>The name is used to identify the service in logs|
|SERVICE_HOST|0.0.0.0|IP address, or a host name that can be resolved to IP addresses|
|SERVICE_PORT|8505|Service port|
|LOG_LEVEL|info|Log level|
|LOG_COLOR_ENABLED|false|Enable the colorized output|
|GRPC_SERVICE_PORT|0|gRPC service port (0 - disabled)|
|GEOIP_DB_PATH|../../db.mmdb|Path to GeoLite2 or GeoIP2 databases|
|CONSUL_ADDRESS||Address of the Consul server|
|CONSUL_SCHEME|http|URI scheme for the Consul server|
|CONSUL_GRPC_SERVICE_ID||The ID of the service. If empty, a random one will be generated|
|CONSUL_GRPC_CLUSTER|grpc_geos|The name of the service to register|
|CONSUL_GRPC_SERVICE_ADDRESS||The address of the service|
|CONSUL_GRPC_SERVICE_PORT|0|The port of the service|
|CONSUL_GRPC_HEALTH_CHECK_TTL|30s|Check TTL|
|CONSUL_GRPC_DEREREGISTER_TTL|30s|If a check is in the critical state for more than this configured value,	then the service will automatically be deregistered|
|CONSUL_REST_SERVICE_ID||The ID of the service. If empty, a random one will be generated|
|CONSUL_REST_CLUSTER|rest_geos|The name of the service to register|
|CONSUL_REST_SERVICE_ADDRESS||The address of the service|
|CONSUL_REST_SERVICE_PORT|0|The port of the service|
|CONSUL_REST_HEALTH_CHECK_TTL|30s|Check TTL|
|CONSUL_REST_DEREREGISTER_TTL|30s|If a check is in the critical state for more than this configured value,	then the service will automatically be deregistered|
|GEONAME_DUMP_DIR||The path to the directory where the GeoNames dumps are located (countryInfo.txt, admin2Codes.txt, cities5000.zip). If variable isn't set, GeoNames api will be disabled. The dumps will be loaded when service starts, if something is missing|
