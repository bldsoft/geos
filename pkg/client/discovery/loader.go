package discovery

import (
	"context"
	"time"

	"github.com/bldsoft/gost/discovery"
	"github.com/jellydator/ttlcache/v3"
)

const clientTTL = 5 * time.Minute

type loader[T any] struct {
	serviceName string
	discovery   discovery.Discovery
	makeClient  func(serviceInfo discovery.ServiceInstanceInfo) (T, error)
	clientCache *ttlcache.Cache[string, []T]
	balancer    discovery.Balancer[T]
}

func newLoader[T any](serviceName string, d discovery.Discovery, makeClient func(serviceInfo discovery.ServiceInstanceInfo) (T, error)) *loader[T] {
	clientCache := ttlcache.New[string, []T]()
	// clientCache.Start()
	return &loader[T]{
		serviceName: serviceName,
		discovery:   d,
		makeClient:  makeClient,
		clientCache: clientCache,
		balancer:    &discovery.RoundRobin[T]{},
	}
}

func (l *loader[T]) Load() ([]T, error) {
	var err error
	if item := l.clientCache.Get(l.serviceName,
		ttlcache.WithDisableTouchOnHit[string, []T](),
		ttlcache.WithLoader[string, []T](ttlcache.LoaderFunc[string, []T](
			func(cache *ttlcache.Cache[string, []T], key string) *ttlcache.Item[string, []T] {
				var serviceInfo *discovery.ServiceInfo
				serviceInfo, err = l.discovery.ServiceByName(context.Background(), l.serviceName)
				if err != nil {
					return nil
				}

				clients := make([]T, 0, len(serviceInfo.Instances))
				for _, info := range serviceInfo.Instances {
					var client T
					if client, err = l.makeClient(info); err != nil {
						return nil
					}
					clients = append(clients, client)
				}

				return l.clientCache.Set(l.serviceName, clients, clientTTL)
			}))); item != nil {
		return l.balancer.Balance(l.serviceName, item.Value()), nil
	}
	return nil, err
}
