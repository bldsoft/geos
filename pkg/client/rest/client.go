package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/bldsoft/geos/pkg/entity"
	"github.com/bldsoft/geos/pkg/microservice"
	"github.com/bldsoft/geos/pkg/storage/geonames"
	"github.com/go-resty/resty/v2"
)

type Client struct {
	client *resty.Client
	apiKey string
}

type RespError struct {
	StatusCode int
	Response   string
}

func (e *RespError) Error() string {
	return fmt.Sprintf("status code: %d, response: %s", e.StatusCode, e.Response)
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
		client: resty.NewWithClient(client).SetBaseURL(baseURL),
	}, nil
}

// only for management endpoints (dump, update, etc.)
func (c *Client) SetApiKey(apiKey string) *Client {
	c.apiKey = apiKey
	return c
}

func (c *Client) APIKey() string {
	return c.apiKey
}

func (c *Client) Origin() string {
	return strings.TrimSuffix(c.client.BaseURL, microservice.BaseApiPath)
}

func get[T any](ctx context.Context, client *resty.Client, path string, query url.Values) (T, error) {
	request := client.R().SetContext(ctx)
	if query != nil {
		request = request.SetQueryParamsFromValues(query)
	}
	return getRequest[T](request, path)
}

func getRequest[T any](req *resty.Request, path string) (T, error) {
	var obj T
	resp, err := req.Get(path)
	if err != nil {
		return obj, err
	}

	if resp.StatusCode() >= 400 {
		return obj, &RespError{StatusCode: resp.StatusCode(), Response: string(resp.Body())}
	}

	err = json.Unmarshal(resp.Body(), &obj)
	if err != nil {
		return obj, err
	}
	return obj, nil
}

func (c *Client) GeoNameContinents(ctx context.Context) []*entity.GeoNameContinent {
	return geonames.GeoNameContinents()
}

func (c *Client) Country(ctx context.Context, address string) (*entity.Country, error) {
	return get[*entity.Country](ctx, c.client, "country/"+address, nil)
}

func (c *Client) City(ctx context.Context, address string, includeISP bool) (*entity.City, error) {
	return get[*entity.City](ctx, c.client, fmt.Sprintf("city/%s?isp=%v", address, includeISP), nil)
}

func (c *Client) CityLite(ctx context.Context, address, lang string) (*entity.CityLite, error) {
	return get[*entity.CityLite](ctx, c.client, "city-lite/"+address, nil)
}

func (c *Client) GeoIPDump(ctx context.Context) (*resty.Response, error) {
	return c.client.R().SetHeader(microservice.APIKey, c.APIKey()).Get("/dump")
}

func getManyWithBody[T any](ctx context.Context, client *resty.Client, path string, body any) ([]*T, error) {
	request := client.R().SetContext(ctx)
	if body != nil {
		request = request.SetBody(body)
	}

	var obj []*T
	resp, err := request.Post(path)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() >= 400 {
		return nil, &RespError{StatusCode: resp.StatusCode(), Response: string(resp.Body())}
	}

	err = json.Unmarshal(resp.Body(), &obj)
	if err != nil {
		return nil, err
	}
	return obj, nil
}

func (c *Client) GeoNameCountries(ctx context.Context, filter entity.GeoNameFilter) ([]*entity.GeoNameCountry, error) {
	return getManyWithBody[entity.GeoNameCountry](ctx, c.client, "geoname/country", filter)
}

func (c *Client) GeoNameSubdivisions(ctx context.Context, filter entity.GeoNameFilter) ([]*entity.GeoNameAdminSubdivision, error) {
	return getManyWithBody[entity.GeoNameAdminSubdivision](ctx, c.client, "geoname/subdivision", filter)
}

func (c *Client) GeoNameCities(ctx context.Context, filter entity.GeoNameFilter) ([]*entity.GeoName, error) {
	return getManyWithBody[entity.GeoName](ctx, c.client, "geoname/city", filter)
}

func (c *Client) GeoNameDump(ctx context.Context, filter entity.GeoNameFilter) (*resty.Response, error) {
	return c.client.R().SetHeader(microservice.APIKey, c.APIKey()).Get("geoname/dump")
}

func (c *Client) requestWithApiKey(ctx context.Context) *resty.Request {
	return c.client.R().SetContext(ctx).SetHeader(microservice.APIKey, c.APIKey())
}

func (c *Client) CheckGeoIPCityUpdates(ctx context.Context) (entity.DBUpdate, error) {
	return getRequest[entity.DBUpdate](c.requestWithApiKey(ctx), "dump/city/update")
}

func (c *Client) CheckGeoIPISPUpdates(ctx context.Context) (entity.DBUpdate, error) {
	return getRequest[entity.DBUpdate](c.requestWithApiKey(ctx), "dump/isp/update")
}

func (c *Client) CheckGeonamesUpdates(ctx context.Context) (entity.DBUpdate, error) {
	return getRequest[entity.DBUpdate](c.requestWithApiKey(ctx), "geoname/update")
}

func (c *Client) UpdateGeoIPCity(ctx context.Context) error {
	_, err := c.requestWithApiKey(ctx).Put("update")
	return err
}

func (c *Client) UpdateGeoIPISP(ctx context.Context) error {
	_, err := c.requestWithApiKey(ctx).Put("dump/isp/update")
	return err
}

func (c *Client) UpdateGeonames(ctx context.Context) error {
	_, err := c.requestWithApiKey(ctx).Put("geoname/update")
	return err
}
