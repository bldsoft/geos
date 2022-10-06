package test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/bldsoft/geos/pkg/client"
	grpc "github.com/bldsoft/geos/pkg/client/grpc"
	rest "github.com/bldsoft/geos/pkg/client/rest"
	"github.com/stretchr/testify/assert"
)

// GEOS_GRPC_ADDRESS=127.0.0.1:8506 GEOS_REST_ADDRESS=127.0.0.1:8505 go test -bench=. -count 1 -benchtime=10000x

var (
	GeosGrpcAddress = os.Getenv("GEOS_GRPC_ADDRESS")
	GeosRestAddress = os.Getenv("GEOS_REST_ADDRESS")
)

type Client struct {
	name   string
	client client.GeoIPClient
}

func clients(b *testing.B) []Client {
	grpc, err := grpc.NewClient(GeosGrpcAddress)
	assert.NoError(b, err)
	rest, err := rest.NewClient(GeosRestAddress)
	assert.NoError(b, err)
	return []Client{
		{"gRPC", grpc},
		{"REST", rest},
	}
}

type ClientRequest struct {
	name   string
	method func(client client.GeoIPClient, address string) (interface{}, error)
}

func requests() []ClientRequest {
	var requests []ClientRequest

	requests = append(requests, ClientRequest{"city", func(client client.GeoIPClient, address string) (interface{}, error) {
		return client.City(context.Background(), address)
	}})
	requests = append(requests, ClientRequest{"country", func(client client.GeoIPClient, address string) (interface{}, error) {
		return client.Country(context.Background(), address)
	}})
	requests = append(requests, ClientRequest{"city-lite", func(client client.GeoIPClient, address string) (interface{}, error) {
		return client.CityLite(context.Background(), address, "")

	}})
	return requests
}

func BenchmarkApi(b *testing.B) {
	for _, parallel := range []bool{false, true} {
		for _, client := range clients(b) {
			for _, req := range requests() {
				for _, addr := range []string{"me", "8.8.8.8"} {
					benchName := fmt.Sprintf("%s_%s_%s", client.name, req.name, addr)
					if parallel {
						b.Run("parallel_"+benchName, func(b *testing.B) {
							b.RunParallel(func(pb *testing.PB) {
								for pb.Next() {
									_, err := req.method(client.client, addr)
									assert.NoError(b, err)
								}
							})
						})
					} else {
						b.Run("serial_"+benchName, func(b *testing.B) {
							for i := 0; i < b.N; i++ {
								_, err := req.method(client.client, addr)
								assert.NoError(b, err)
							}
						})
					}
				}
			}
		}
	}
}
