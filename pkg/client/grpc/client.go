package client

import (
	"context"

	mapping "github.com/bldsoft/geos/pkg/controller/grpc"
	"google.golang.org/grpc"

	pb "github.com/bldsoft/geos/pkg/controller/grpc/proto"
	"github.com/oschwald/geoip2-golang"
)

type Client struct {
	conn   *grpc.ClientConn
	client pb.GeoServiceClient
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
		client: pb.NewGeoServiceClient(conn),
	}, nil
}

func (c *Client) Country(ctx context.Context, ip string) (*geoip2.Country, error) {
	country, err := c.client.Country(ctx, &pb.IpRequest{Ip: ip})
	if err != nil {
		return nil, err
	}
	return mapping.PbToCountry(country), nil
}

func (c *Client) City(ctx context.Context, ip string) (*geoip2.City, error) {
	city, err := c.client.City(ctx, &pb.IpRequest{Ip: ip})
	if err != nil {
		return nil, err
	}
	return mapping.PbToCity(city), nil
}

func (c *Client) Close() {
	c.conn.Close()
}
