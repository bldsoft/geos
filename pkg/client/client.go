package client

import (
	"context"
	"errors"
	"time"

	grpc_client "github.com/bldsoft/geos/pkg/client/grpc"
	rest_client "github.com/bldsoft/geos/pkg/client/rest"
)

type Opt func(Client)

func WithApiKey(apiKey string) Opt {
	return func(c Client) {
		restClient, ok := c.(*rest_client.Client)
		if ok {
			restClient.SetApiKey(apiKey)
		}
	}
}

func NewClientWithOpt(addrs []string, opts ...Opt) (Client, error) {
	res := &MultiClient{}
	var multiErr error

	for _, addr := range addrs {
		if client, err := newClient(addr, ""); err == nil {
			for _, opt := range opts {
				opt(client)
			}
			res.Clients = append(res.Clients, client)
		} else {
			multiErr = errors.Join(multiErr, err)
		}
	}

	if len(res.Clients) == 0 {
		return nil, multiErr
	}
	return res, nil

}

func NewClient(addrs ...string) (Client, error) {
	return NewClientWithOpt(addrs)
}

func newClient(addr string, apiKey string) (Client, error) {
	grpcClient := func(addr string) (Client, error) {
		return grpc_client.NewClient(addr)
	}
	restClient := func(addr string) (Client, error) {
		return rest_client.NewClient(addr)
	}

	var multiErr error
	for _, getClient := range []func(addr string) (Client, error){
		grpcClient,
		restClient,
	} {
		client, err := getClient(addr)
		if err == nil {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_, err = client.CityLite(ctx, "8.8.8.8", "")
			if err == nil {
				return client, nil
			}
		}
		multiErr = errors.Join(multiErr, err)
	}
	return restClient(addr)
}
