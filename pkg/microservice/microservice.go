package microservice

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"github.com/bldsoft/geos/pkg/config"
	"github.com/bldsoft/geos/pkg/controller"
	"github.com/bldsoft/geos/pkg/controller/rest"
	"github.com/bldsoft/geos/pkg/repository"
	"github.com/bldsoft/geos/pkg/service"
	"github.com/bldsoft/geos/pkg/storage/geonames"
	"github.com/bldsoft/gost/auth"
	"github.com/bldsoft/gost/clickhouse"
	gost "github.com/bldsoft/gost/controller"
	"github.com/bldsoft/gost/discovery"
	"github.com/bldsoft/gost/discovery/common"
	"github.com/bldsoft/gost/discovery/inhouse"
	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/server"
	gost_storage "github.com/bldsoft/gost/storage"
	"github.com/go-chi/chi/v5"
)

const (
	BaseApiPath                 = "/geoip"
	APIKey                      = "GEOS-API-Key"
	APIKeyMetaKey               = "api-key"
	MMDBCitiesBuildEpochMetaKey = "mmdbCityTs"
	MMDBIspBuildEpochMetaKey    = "mmdbISPTs"
	GrpcAddressMetaKey          = "grpc-address"
	ServiceName                 = config.ServiceName
)

type Microservice struct {
	config *config.Config

	geoIpService   controller.GeoIpService
	geoNameService controller.GeoNameService

	discovery discovery.Discovery

	asyncRunners []server.AsyncRunner
}

func NewMicroservice(config config.Config) *Microservice {
	srv := &Microservice{
		config: &config,
	}
	srv.initServices()
	return srv
}

func (srv *Microservice) getDbJobGroup(db gost_storage.IStorage) *server.AsyncJobGroup {
	dbAsyncJobs := server.NewAsyncJobGroup()
	dbAsyncJobChain := server.NewAsyncJobChain(server.NewAsyncJob(nil, db.Disconnect), dbAsyncJobs)
	srv.asyncRunners = append(srv.asyncRunners, dbAsyncJobChain)
	return dbAsyncJobs
}

func (m *Microservice) setDiscoveryMeta() {
	if ispMeta, err := m.geoIpService.MetaData(context.Background(), repository.MaxmindDBTypeISP); err == nil {
		m.discovery.SetMetadata(MMDBIspBuildEpochMetaKey, fmt.Sprintf("%d", ispMeta.BuildEpoch))
	} else {
		log.Logger.ErrorWithFields(log.Fields{"err": err}, "failed to get isp database metadata")
	}

	if citiesMeta, err := m.geoIpService.MetaData(context.Background(), repository.MaxmindDBTypeCity); err == nil {
		m.discovery.SetMetadata(MMDBCitiesBuildEpochMetaKey, fmt.Sprintf("%d", citiesMeta.BuildEpoch))
	} else {
		log.Logger.ErrorWithFields(log.Fields{"err": err}, "failed to get cities database metadata")
	}

	m.discovery.SetMetadata(APIKeyMetaKey, m.config.ApiKey)
}

func (m *Microservice) initServices() {
	if len(m.config.Clickhouse.Dsn) != 0 {
		var wg sync.WaitGroup
		clickhouseDB := clickhouse.NewStorage(m.config.Clickhouse)
		gost_storage.DBConnectAsync(&wg, clickhouseDB.Connect, -1, time.Second)
		wg.Wait()

		logExporter := clickhouse.NewLogExporter(clickhouseDB, m.config.LogExport)
		log.Logger.AddExporter(logExporter, nil)
		m.getDbJobGroup(clickhouseDB).Append(logExporter)
	} else {
		log.Debug("Log export to ClickHouse is off")
	}

	rep := repository.NewGeoIPRepository(m.config.GeoDbPath, m.config.GeoDbISPPath, m.config.GeoIPCsvDumpDirPath)
	m.geoIpService = service.NewGeoIpService(rep)

	geoNameStorage := m.geonamesStorage()
	geoNameRep := repository.NewGeoNamesRepository(geoNameStorage)
	m.geoNameService = service.NewGeoNameService(geoNameRep)

	m.discovery = common.NewDiscovery(m.config.Server, m.config.Discovery)
	m.setDiscoveryMeta()

	m.asyncRunners = append(m.asyncRunners, m.discovery)

	if m.config.NeedGrpc() {
		grpcService := NewGrpcMicroservice(m.config.GRPCServiceBindAddress.HostPort(), m.geoIpService, m.geoNameService)
		m.asyncRunners = append(m.asyncRunners, grpcService)

		m.discovery.SetMetadata(GrpcAddressMetaKey, m.config.GRPCServiceAddress.String())
	} else {
		log.Info("gRPC is off")
	}

}

func (m *Microservice) geonamesStorage() geonames.Storage {
	original := geonames.NewStorage(m.config.GeoNameDumpDirPath)
	customs := geonames.NewCustomStoragesFromDir(filepath.Dir(m.config.GeoDbPath), "geonames")
	return geonames.NewMultiStorage(original).Add(customs...)
}

func (m *Microservice) BuildRoutes(router chi.Router) {
	if d, ok := m.discovery.(*inhouse.Discovery); ok {
		d.Mount(router)
	}
	router.Route(BaseApiPath, func(r chi.Router) {
		r.Get("/ping", gost.GetPingHandler)
		r.With(m.ApiKeyMiddleware()).Get("/env", gost.GetEnvHandler(m.config, nil))
		r.Get("/version", gost.GetVersionHandler)

		geoIpController := rest.NewGeoIpController(m.geoIpService)
		r.Get("/country/{addr}", geoIpController.GetCountryHandler)
		r.Get("/city/{addr}", geoIpController.GetCityHandler)
		r.Get("/city-lite/{addr}", geoIpController.GetCityLiteHandler)

		r.With(m.ApiKeyMiddleware()).Get("/dump", geoIpController.GetDumpHandler) // deprecated, used by streampool
		r.Route("/dump/{db}", func(r chi.Router) {
			r.Use(m.ApiKeyMiddleware())
			r.Get("/csv", geoIpController.GetCSVDatabaseHandler)
			r.Get("/mmdb", geoIpController.GetMMDBDatabaseHandler)
			r.Get("/metadata", geoIpController.GetDatabaseMetaHandler)
		})

		geoNameController := rest.NewGeoNameController(m.geoNameService)
		r.Route("/geoname", func(r chi.Router) {
			r.Get("/continent", geoNameController.GetGeoNameContinentsHandler)
			r.Get("/country", geoNameController.GetGeoNameCountriesHandler)
			r.Post("/country", geoNameController.GetGeoNameCountriesHandler)
			r.Get("/subdivision", geoNameController.GetGeoNameSubdivisionsHandler)
			r.Post("/subdivision", geoNameController.GetGeoNameSubdivisionsHandler)
			r.Get("/city", geoNameController.GetGeoNameCitiesHandler)
			r.Post("/city", geoNameController.GetGeoNameCitiesHandler)
			r.With(m.ApiKeyMiddleware()).Get("/dump", geoNameController.GetDumpHandler)
		})
	})
}

func (m *Microservice) ApiKeyMiddleware() func(next http.Handler) http.Handler {
	return auth.ApiKeyMiddleware(APIKey, m.config.ApiKey)
}

func (m *Microservice) GetAsyncRunners() []server.AsyncRunner {
	return m.asyncRunners
}
