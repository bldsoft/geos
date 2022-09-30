package grpc

import (
	context "context"
	"net"
	"strings"

	"github.com/bldsoft/geos/pkg/controller"
	pb "github.com/bldsoft/geos/pkg/controller/grpc/proto"
	"github.com/bldsoft/gost/log"
	"google.golang.org/grpc/peer"
)

//go:generate protoc -I=../../.. --go_out=proto --go-grpc_out=proto api/grpc/geoip.proto

type GeoIpController struct {
	pb.UnimplementedGeoIpServiceServer
	service controller.GeoIpService
}

func NewGeoIpController(geoService controller.GeoIpService) *GeoIpController {
	return &GeoIpController{service: geoService}
}

func (c *GeoIpController) ip(ctx context.Context, req *pb.IpRequest) net.IP {
	if req.GetIp() == "me" {
		p, _ := peer.FromContext(ctx)
		return net.ParseIP(strings.Split(p.Addr.String(), ":")[0])
	}
	return net.ParseIP(req.GetIp())
}

func (c *GeoIpController) Country(ctx context.Context, req *pb.IpRequest) (*pb.CountryResponse, error) {
	country, err := c.service.Country(ctx, c.ip(ctx, req))
	if err != nil {
		log.FromContext(ctx).Error(err.Error())
		return nil, err
	}
	return CountryToPb(country), nil
}

func (c *GeoIpController) City(ctx context.Context, req *pb.IpRequest) (*pb.CityResponse, error) {
	city, err := c.service.City(ctx, c.ip(ctx, req))
	if err != nil {
		log.FromContext(ctx).Error(err.Error())
		return nil, err
	}
	return CityToPb(city), nil

}
