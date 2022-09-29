package grpc

import (
	context "context"
	"net"

	"github.com/bldsoft/geos/pkg/controller"
	proto "github.com/bldsoft/geos/pkg/controller/grpc/proto"
	"github.com/bldsoft/gost/log"
)

//go:generate protoc --go_out=. --go-grpc_out=. proto/geoip.proto

type GeoController struct {
	proto.UnimplementedGeoServiceServer
	service controller.GeoService
}

func NewGeoController(geoService controller.GeoService) *GeoController {
	return &GeoController{service: geoService}
}

func (c *GeoController) Country(ctx context.Context, req *proto.IpRequest) (*proto.CountryResponse, error) {
	country, err := c.service.Country(ctx, net.ParseIP(req.GetIp()))
	if err != nil {
		log.FromContext(ctx).Error(err.Error())
		return nil, err
	}
	return CountryToPb(country), nil
}

func (c *GeoController) City(ctx context.Context, req *proto.IpRequest) (*proto.CityResponse, error) {
	city, err := c.service.City(ctx, net.ParseIP(req.GetIp()))
	if err != nil {
		log.FromContext(ctx).Error(err.Error())
		return nil, err
	}
	return CityToPb(city), nil

}
