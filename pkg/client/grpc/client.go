package client

import (
	"context"

	mapping "github.com/bldsoft/geos/pkg/controller/grpc"
	"github.com/bldsoft/geos/pkg/entity"
	"github.com/bldsoft/geos/pkg/microservice/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	pb "github.com/bldsoft/geos/pkg/controller/grpc/proto"
)

type Client struct {
	conn   *grpc.ClientConn
	client pb.GeoIpServiceClient
}

func NewClient(addr string) (*Client, error) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	conn, err := grpc.Dial(addr, opts...)
	if err != nil {
		return nil, err
	}
	return &Client{
		conn:   conn,
		client: pb.NewGeoIpServiceClient(conn),
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
	country, err := c.client.Country(ctx, &pb.AddrRequest{Address: address})
	if err != nil {
		return nil, err
	}
	return mapping.PbToCountry(country), nil
}

func (c *Client) City(ctx context.Context, address string) (*entity.City, error) {
	ctx = c.prepareContext(ctx)
	city, err := c.client.City(ctx, &pb.AddrRequest{Address: address})
	if err != nil {
		return nil, err
	}
	return mapping.PbToCity(city), nil
}

func (c *Client) CityLite(ctx context.Context, address, lang string) (*entity.CityLite, error) {
	ctx = c.prepareContext(ctx)
	cityLite, err := c.client.CityLite(ctx, &pb.CityLiteRequest{Address: address, Lang: lang})
	if err != nil {
		return nil, err
	}
	return mapping.PbToCityLite(cityLite), nil
}

func (c *Client) Close() {
	c.conn.Close()
}
