package consul

import (
	"fmt"
	"time"

	"github.com/bldsoft/gost/consul"
	"github.com/hashicorp/consul/api"
	"github.com/jellydator/ttlcache/v3"
)

const clientTTL = 5 * time.Minute

type loader[T any] struct {
	serviceCluster string
	discovery      *consul.Discovery
	makeClient     func(ci api.AgentServiceChecksInfo) (T, error)
	clientCache    *ttlcache.Cache[string, []T]
	balancer       consul.Balancer[T]
}

func newLoader[T any](serviceCluster string, discovery *consul.Discovery, makeClient func(ci api.AgentServiceChecksInfo) (T, error)) *loader[T] {
	clientCache := ttlcache.New[string, []T]()
	// clientCache.Start()
	return &loader[T]{
		serviceCluster: serviceCluster,
		discovery:      discovery,
		makeClient:     makeClient,
		clientCache:    clientCache,
		balancer:       &consul.RoundRobin[T]{},
	}
}

func (l *loader[T]) Load() ([]T, error) {
	var err error
	if item := l.clientCache.Get(l.serviceCluster,
		ttlcache.WithDisableTouchOnHit[string, []T](),
		ttlcache.WithLoader[string, []T](ttlcache.LoaderFunc[string, []T](
			func(cache *ttlcache.Cache[string, []T], key string) *ttlcache.Item[string, []T] {
				var infos []api.AgentServiceChecksInfo
				_, infos, err = l.discovery.ApiClient().Agent().AgentHealthServiceByName(l.serviceCluster)
				if err != nil {
					return nil
				}
				if len(infos) == 0 {
					err = fmt.Errorf("discovery: service %s not found", l.serviceCluster)
					return nil
				}

				clients := make([]T, 0, len(infos))
				for _, info := range infos {
					var client T
					if client, err = l.makeClient(info); err != nil {
						return nil
					}
					clients = append(clients, client)
				}

				return l.clientCache.Set(l.serviceCluster, clients, clientTTL)
			}))); item != nil {
		return l.balancer.Balance(l.serviceCluster, item.Value()), nil
	}
	return nil, err
}
