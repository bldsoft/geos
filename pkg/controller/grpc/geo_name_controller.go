package grpc

import (
	"github.com/bldsoft/geos/pkg/controller"
	pb "github.com/bldsoft/geos/pkg/controller/grpc/proto"
	"github.com/bldsoft/geos/pkg/entity"
	"github.com/bldsoft/gost/log"
)

//go:generate protoc -I=../../.. --go_out=proto --go-grpc_out=proto api/grpc/geoname.proto

type GeoNameController struct {
	pb.UnimplementedGeoNameServiceServer
	service controller.GeoNameService
}

func NewGeoNameController(geoNameService controller.GeoNameService) *GeoNameController {
	return &GeoNameController{service: geoNameService}
}

func sendToStream[T any, P any](objects []*T, convert func(*T) *P, stream interface {
	Send(*P) error
}) error {
	for _, obj := range objects {
		if err := stream.Send(convert(obj)); err != nil {
			return err
		}
	}
	return nil
}

func (c *GeoNameController) Country(_ *pb.GeoNameRequest, stream pb.GeoNameService_CountryServer) error {
	ctx := stream.Context()
	countries, err := c.service.Countries(ctx)
	if err != nil {
		log.FromContext(ctx).Error(err.Error())
		return err
	}
	return sendToStream[entity.GeoNameCountry, pb.GeoNameCountryResponse](countries, GeoNameCountryToPb, stream)
}

func (c *GeoNameController) Subdivision(_ *pb.GeoNameRequest, stream pb.GeoNameService_SubdivisionServer) error {
	ctx := stream.Context()
	subdivisions, err := c.service.Subdivisions(ctx)
	if err != nil {
		log.FromContext(ctx).Error(err.Error())
		return err
	}
	return sendToStream[entity.AdminSubdivision, pb.GeoNameSubdivisionResponse](subdivisions, GeoNameSubdivisionToPb, stream)
}

func (c *GeoNameController) City(_ *pb.GeoNameRequest, stream pb.GeoNameService_CityServer) error {
	ctx := stream.Context()
	cities, err := c.service.Cities(ctx)
	if err != nil {
		log.FromContext(ctx).Error(err.Error())
		return err
	}
	return sendToStream[entity.Geoname, pb.GeoNameCityResponse](cities, GeoNameCityToPb, stream)
}
