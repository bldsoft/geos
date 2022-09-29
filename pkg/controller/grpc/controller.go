package grpc

import (
	context "context"
	"net"

	"github.com/bldsoft/geos/pkg/controller"
	pb "github.com/bldsoft/geos/pkg/controller/grpc/proto"
	"github.com/bldsoft/gost/log"
)

type GeoIpController struct {
	pb.UnimplementedGeoIpServiceServer
	service controller.GeoIpService
}

func NewGeoIpController(geoService controller.GeoIpService) *GeoIpController {
	return &GeoIpController{service: geoService}
}

func (c *GeoIpController) Country(ctx context.Context, req *pb.IpRequest) (*pb.CountryResponse, error) {
	country, err := c.service.Country(ctx, net.ParseIP(req.GetIp()))
	if err != nil {
		log.FromContext(ctx).Error(err.Error())
		return nil, err
	}
	return CountryToPb(country), nil
}

func (c *GeoIpController) City(ctx context.Context, req *pb.IpRequest) (*pb.CityResponse, error) {
	city, err := c.service.City(ctx, net.ParseIP(req.GetIp()))
	if err != nil {
		log.FromContext(ctx).Error(err.Error())
		return nil, err
	}
	return CityToPb(city), nil

}
