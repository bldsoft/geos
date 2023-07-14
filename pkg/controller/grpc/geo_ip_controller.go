package grpc

import (
	context "context"

	"github.com/bldsoft/geos/pkg/controller"
	pb "github.com/bldsoft/geos/pkg/controller/grpc/proto"
	"github.com/bldsoft/gost/log"
)

//go:generate protoc -I=../../.. --go_out=proto --go-grpc_out=proto api/grpc/geoip.proto

type GeoIpController struct {
	pb.UnimplementedGeoIpServiceServer
	service controller.GeoIpService
}

func NewGeoIpController(geoService controller.GeoIpService) *GeoIpController {
	return &GeoIpController{service: geoService}
}

func (c *GeoIpController) Country(ctx context.Context, req *pb.AddrRequest) (*pb.CountryResponse, error) {
	country, err := c.service.Country(ctx, req.Address)
	if err != nil {
		log.FromContext(ctx).Error(err.Error())
		return nil, err
	}
	return CountryToPb(country), nil
}

func (c *GeoIpController) City(ctx context.Context, req *pb.AddrRequest) (*pb.CityResponse, error) {
	city, err := c.service.City(ctx, req.Address, false)
	if err != nil {
		log.FromContext(ctx).Error(err.Error())
		return nil, err
	}
	return CityToPb(city), nil
}

func (c *GeoIpController) CityLite(ctx context.Context, req *pb.CityLiteRequest) (*pb.CityLiteResponse, error) {
	cityLite, err := c.service.CityLite(ctx, req.Address, req.Lang)
	if err != nil {
		log.FromContext(ctx).Error(err.Error())
		return nil, err
	}
	return CityLiteToPb(cityLite), nil
}
