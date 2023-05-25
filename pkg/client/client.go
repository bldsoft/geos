package client

import (
	"errors"

	grpc_client "github.com/bldsoft/geos/pkg/client/grpc"
	rest_client "github.com/bldsoft/geos/pkg/client/rest"
)

func NewClient(addr string) (Client, error) {
	var multiErr error
	grpcClient, err := grpc_client.NewClient(addr)
	if err == nil {
		return grpcClient, nil
	}
	multiErr = errors.Join(multiErr, err)

	restClient, err := rest_client.NewClient(addr)
	if err == nil {
		return restClient, nil
	}

	return nil, errors.Join(multiErr, err)
}
