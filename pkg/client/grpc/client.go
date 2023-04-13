package client

import (
	"context"
	"errors"
	"io"

	mapping "github.com/bldsoft/geos/pkg/controller/grpc"
	"github.com/bldsoft/geos/pkg/entity"
	"github.com/bldsoft/geos/pkg/microservice/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	pb "github.com/bldsoft/geos/pkg/controller/grpc/proto"
)

type Client struct {
	conn          *grpc.ClientConn
	geoIpClient   pb.GeoIpServiceClient
	geoNameClient pb.GeoNameServiceClient
}

func NewClient(addr string, opts ...grpc.DialOption) (*Client, error) {
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	conn, err := grpc.Dial(addr, opts...)
	if err != nil {
		return nil, err
	}
	return &Client{
		conn:          conn,
		geoIpClient:   pb.NewGeoIpServiceClient(conn),
		geoNameClient: pb.NewGeoNameServiceClient(conn),
	}, nil
}

func (c *Client) prepareContext(ctx context.Context) context.Context {
	if reqID := middleware.GetReqID(ctx); reqID != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, middleware.RequestIDHeader, reqID)
	}
	if realIP := middleware.GetRealIP(ctx); realIP != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, middleware.RealIPHeader, realIP)
	}
	return ctx
}

func (c *Client) Country(ctx context.Context, address string) (*entity.Country, error) {
	ctx = c.prepareContext(ctx)
	country, err := c.geoIpClient.Country(ctx, &pb.AddrRequest{Address: address})
	if err != nil {
		return nil, err
	}
	return mapping.PbToCountry(country), nil
}

func (c *Client) City(ctx context.Context, address string) (*entity.City, error) {
	ctx = c.prepareContext(ctx)
	city, err := c.geoIpClient.City(ctx, &pb.AddrRequest{Address: address})
	if err != nil {
		return nil, err
	}
	return mapping.PbToCity(city), nil
}

func (c *Client) CityLite(ctx context.Context, address, lang string) (*entity.CityLite, error) {
	ctx = c.prepareContext(ctx)
	cityLite, err := c.geoIpClient.CityLite(ctx, &pb.CityLiteRequest{Address: address, Lang: lang})
	if err != nil {
		return nil, err
	}
	return mapping.PbToCityLite(cityLite), nil
}

func recvAll[R, T any](stream interface {
	Recv() (*R, error)
}, convert func(*R) *T) ([]*T, error) {
	var res []*T
	for {
		resp, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		res = append(res, convert(resp))
	}
	return res, nil
}

func (c *Client) GeoNameCountries(ctx context.Context) ([]*entity.GeoNameCountry, error) {
	ctx = c.prepareContext(ctx)
	countryClient, err := c.geoNameClient.Country(ctx, &pb.GeoNameRequest{})
	if err != nil {
		return nil, err
	}
	return recvAll[pb.GeoNameCountryResponse](countryClient, mapping.PbToGeoNameCountry)
}

func (c *Client) GeoNameSubdivisions(ctx context.Context) ([]*entity.AdminSubdivision, error) {
	ctx = c.prepareContext(ctx)
	subdivisionClient, err := c.geoNameClient.Subdivision(ctx, &pb.GeoNameRequest{})
	if err != nil {
		return nil, err
	}
	return recvAll[pb.GeoNameSubdivisionResponse](subdivisionClient, mapping.PbToGeoNameSubdivision)
}

func (c *Client) GeoNameCities(ctx context.Context) ([]*entity.Geoname, error) {
	ctx = c.prepareContext(ctx)
	cityClient, err := c.geoNameClient.City(ctx, &pb.GeoNameRequest{})
	if err != nil {
		return nil, err
	}
	return recvAll[pb.GeoNameCityResponse](cityClient, mapping.PbToGeoNameCity)
}

func (c *Client) Close() {
	c.conn.Close()
}
