|**Environment variable**|**Value**|**Description**|
|------------------------|---------|---------------|
|SERVICE_NAME|hostname|DEPRECATED. Unique service instance name. Use 'hostname' to set the hostname value. <br/>DEPRECATED. The name is used to identify the service in logs. |
|SERVICE_INSTANCE_NAME||Unique service instance name. Use 'hostname' to set the hostname value. <br/>The name is used to identify the service in logs. |
|SERVICE_HOST|0.0.0.0|DEPRECATED. IP address, or a host name that can be resolved to IP addresses|
|SERVICE_PORT|8505|DEPRECATED. Service port|
|SERVICE_BIND_ADDRESS||Service configuration related to what address bind to and port to listen on|
|SERVICE_ADDRESS||Service public address|
|LOG_LEVEL|info|Log level|
|LOG_COLOR_ENABLED|false|Enable the colorized output|
|CLICKHOUSE_DSN||Clickhouse DSN|
|CLICKHOUSE_LOG_EXPORT_FLUSH_TIME_MS|1000|Max time between log exporting|
|CLICKHOUSE_LOG_EXPORT_MAX_BATCH_SIZE|1000|Max batch size for log insert query|
|CLICKHOUSE_LOG_EXPORT_TABLE|LOG_RECORDS|Table name for log exporting|
|GRPC_SERVICE_BIND_ADDRESS|0.0.0.0:8506|Service configuration related to what address bind to and port to listen|
|GRPC_SERVICE_ADDRESS|0.0.0.0:8506|GRPC public address|
|GEOIP_DB_PATH|../../db.mmdb|Path to GeoLite2 or GeoIP2 city database|
|GEOIP_DB_ISP_PATH||Path to GeoIP2 ISP database|
|DISCOVERY_TYPE|none|Discovery type (none, in-house, consul)|
|DISCOVERY_INHOUSE_EMBEDDED|true|If true, in-house discovery will use service bind address|
|DISCOVERY_INHOUSE_BIND_ADDRESS|0.0.0.0:3001|For non embedded mode. Configuration related to what address to bind to and ports to listen on.|
|DISCOVERY_INHOUSE_CLUSTER_MEMBERS||Comma separated list of any existing member of the cluster to join it. Example: '127.0.0.1:3001'|
|DISCOVERY_INHOUSE_SECRET_KEY|ZljFlK6atNj5U3VbHrDxRgFMHYcgEOpy|SecretKey is used to encrypt messages. The value should be either 16, 24, or 32 bytes to select AES-128, AES-192, or AES-256.|
|DISCOVERY_CONSUL_ADDRESS|http://127.0.0.1:8500|Address of the Consul server|
|DISCOVERY_CONSUL_TOKEN|| Token is used to provide a per-request ACL token|
|DISCOVERY_CONSUL_HEALTH_CHECK_TTL|30s|Check TTL|
|DISCOVERY_CONSUL_DEREREGISTER_TTL|30s|If a check is in the critical state for more than this configured value,	then the service will automatically be deregistered|
|GEONAME_DUMP_DIR||The path to the directory where the GeoNames dumps are located (countryInfo.txt, admin1CodesASKII.txt, cities5000.zip). If variable isn't set, GeoNames api will be disabled. The dumps will be loaded when service starts, if something is missing|
|GEOIP_DUMP_DIR||The path to the directory where the csv ip database is located. If the variable is set and the csv file is missing, the service will generate it from the mmdb when it starts.|
|API_KEY|Dfga4pBfeRsMnxesWmY8eNBCW2Zf46kL|API key for dumps used for importing into other databases|
