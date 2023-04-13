package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	grpc "github.com/bldsoft/geos/pkg/client/grpc"
	"github.com/bldsoft/geos/pkg/config"
	"github.com/urfave/cli/v2"

	gost "github.com/bldsoft/gost/config"
)

var (
	host string
	port uint
)

func client(ctx *cli.Context) *grpc.Client {
	addr := fmt.Sprintf("%s:%d", host, port)
	client, err := grpc.NewClient(addr)
	if err != nil {
		log.Fatalf(err.Error())
	}
	return client
}

func addr(ctx *cli.Context) string {
	addr := ctx.Args().Get(0)
	if len(addr) == 0 {
		return "me"
	}
	return addr
}

func commonFlags() []cli.Flag {
	var defaults config.Config
	_ = gost.SetDefaults(&defaults)
	return []cli.Flag{
		&cli.StringFlag{
			Name:        "host",
			Usage:       "Service host",
			Value:       defaults.Server.Host,
			Destination: &host,
			Aliases:     []string{"H"},
		},
		&cli.UintFlag{
			Name:        "port",
			Usage:       "Service gRPC port",
			Value:       uint(defaults.GrpcPort),
			Destination: &port,
			Aliases:     []string{"p"},
		},
	}
}

func print(obj interface{}) error {
	data, err := json.MarshalIndent(obj, "", "	")
	if err != nil {
		return err
	}
	fmt.Printf("%s", data)
	return nil
}

func main() {
	app := &cli.App{
		Name:  "geos-cli",
		Usage: "Geos gRPC client",
		Flags: commonFlags(),
		Commands: []*cli.Command{
			{
				Name: "city",
				Action: func(ctx *cli.Context) error {
					city, err := client(ctx).City(ctx.Context, addr(ctx))
					if err != nil {
						return err
					}
					return print(city)
				},
			},
			{
				Name: "country",
				Action: func(ctx *cli.Context) error {
					country, err := client(ctx).Country(ctx.Context, addr(ctx))
					if err != nil {
						return err
					}
					return print(country)
				},
			},
			{
				Name: "city-lite",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "lang",
						Value:   "en",
						Usage:   "Language for country 1and city name",
						Aliases: []string{"l"}},
				},
				Action: func(ctx *cli.Context) error {
					cityLite, err := client(ctx).CityLite(ctx.Context, addr(ctx), ctx.String("lang"))
					if err != nil {
						return err
					}
					return print(cityLite)
				},
			},
			{
				Name: "geoname-country",
				Action: func(ctx *cli.Context) error {
					countries, err := client(ctx).GeoNameCountries(ctx.Context)
					if err != nil {
						return err
					}
					return print(countries)
				},
			},
			{
				Name: "geoname-subdivision",
				Action: func(ctx *cli.Context) error {
					subdivisions, err := client(ctx).GeoNameSubdivisions(ctx.Context)
					if err != nil {
						return err
					}
					return print(subdivisions)
				},
			},
			{
				Name: "geoname-city",
				Action: func(ctx *cli.Context) error {
					cities, err := client(ctx).GeoNameCities(ctx.Context)
					if err != nil {
						return err
					}
					return print(cities)
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
