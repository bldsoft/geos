package client

import (
	"context"

	mapping "github.com/bldsoft/geos/pkg/controller/grpc"
	"github.com/bldsoft/geos/pkg/entity"
	"google.golang.org/grpc"

	pb "github.com/bldsoft/geos/pkg/controller/grpc/proto"
)

type Client struct {
	conn   *grpc.ClientConn
	client pb.GeoIpServiceClient
}

func NewClient(addr string) (*Client, error) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.Dial(addr, opts...)
	if err != nil {
		return nil, err
	}
	return &Client{
		conn:   conn,
		client: pb.NewGeoIpServiceClient(conn),
	}, nil
}

func (c *Client) Country(ctx context.Context, ip string) (*entity.Country, error) {
	country, err := c.client.Country(ctx, &pb.IpRequest{Ip: ip})
	if err != nil {
		return nil, err
	}
	return mapping.PbToCountry(country), nil
}

func (c *Client) City(ctx context.Context, ip string) (*entity.City, error) {
	city, err := c.client.City(ctx, &pb.IpRequest{Ip: ip})
	if err != nil {
		return nil, err
	}
	return mapping.PbToCity(city), nil
}

func (c *Client) CityLite(ctx context.Context, ip, lang string) (*entity.CityLite, error) {
	cityLite, err := c.client.CityLite(ctx, &pb.CityLiteRequest{Ip: ip, Lang: lang})
	if err != nil {
		return nil, err
	}
	return mapping.PbToCityLite(cityLite), nil
}

func (c *Client) Close() {
	c.conn.Close()
}
