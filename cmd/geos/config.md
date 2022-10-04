| **Environment variable** | **Value**     | **Description**                                                                   |
| ------------------------ | ------------- | --------------------------------------------------------------------------------- |
| SERVICE_NAME             | default_name  | Unique service instanse name<br/>The name is used to identify the service in logs |
| SERVICE_HOST             | 0.0.0.0       | IP address, or a host name that can be resolved to IP addresses                   |
| SERVICE_PORT             | 8505          | Service port                                                                      |
| LOG_LEVEL                | info          | Log level                                                                         |
| LOG_COLOR_ENABLED        | true          | Enable the colorized output                                                       |
| GRPC_SERVICE_PORT        | 0             | gRPC service port (0 - disabled)                                                  |
| GEOIP_DB_PATH            | ../../db.mmdb | Path to GeoLite2 or GeoIP2 databases                                              |
