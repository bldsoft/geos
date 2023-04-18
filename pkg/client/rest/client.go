package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/bldsoft/geos/pkg/entity"
	"github.com/bldsoft/geos/pkg/microservice"
	"github.com/bldsoft/geos/pkg/storage"
	"github.com/bldsoft/gost/utils"
	"github.com/go-resty/resty/v2"
)

type Client struct {
	client *resty.Client
}

func NewClient(addr string) (*Client, error) {
	return NewWithClient(addr, http.DefaultClient)
}

func NewWithClient(addr string, client *http.Client) (*Client, error) {
	if !strings.HasPrefix(addr, "http://") && !strings.HasPrefix(addr, "https://") {
		addr = "http://" + addr
	}
	baseURL, err := url.JoinPath(addr, microservice.BaseApiPath)
	if err != nil {
		return nil, err
	}
	return &Client{
		resty.NewWithClient(client).SetBaseURL(baseURL),
	}, nil
}

func get[T any](ctx context.Context, client *resty.Client, path string, query url.Values) (*T, error) {
	request := client.R().SetContext(ctx)
	if query != nil {
		request = request.SetQueryParamsFromValues(query)
	}
	var obj T
	resp, err := request.Get(path)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(resp.Body(), &obj)
	if err != nil {
		return nil, err
	}
	return &obj, nil
}

func (c *Client) GeoNameContinents(ctx context.Context) []*entity.GeoNameContinent {
	return storage.GeoNameContinents()
}

func (c *Client) Country(ctx context.Context, address string) (*entity.Country, error) {
	return get[entity.Country](ctx, c.client, "country/"+address, nil)
}

func (c *Client) City(ctx context.Context, address string) (*entity.City, error) {
	return get[entity.City](ctx, c.client, "city/"+address, nil)
}

func (c *Client) CityLite(ctx context.Context, address, lang string) (*entity.CityLite, error) {
	return get[entity.CityLite](ctx, c.client, "city-lite/"+address, nil)
}

func getMany[T any](ctx context.Context, client *resty.Client, path string, query url.Values) ([]*T, error) {
	request := client.R().SetContext(ctx)
	if query != nil {
		request = request.SetQueryParamsFromValues(query)
	}
	var obj []*T
	resp, err := request.Get(path)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(resp.Body(), &obj)
	if err != nil {
		return nil, err
	}
	return obj, nil
}

func (c *Client) GeoNameCountries(ctx context.Context, filter entity.GeoNameFilter) ([]*entity.GeoNameCountry, error) {
	return getMany[entity.GeoNameCountry](ctx, c.client, "geoname/country", utils.Query(filter))
}

func (c *Client) GeoNameSubdivisions(ctx context.Context, filter entity.GeoNameFilter) ([]*entity.AdminSubdivision, error) {
	return getMany[entity.AdminSubdivision](ctx, c.client, "geoname/subdivision", utils.Query(filter))
}

func (c *Client) GeoNameCities(ctx context.Context, filter entity.GeoNameFilter) ([]*entity.Geoname, error) {
	return getMany[entity.Geoname](ctx, c.client, "geoname/city", utils.Query(filter))
}
