package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	grpc "github.com/bldsoft/geos/pkg/client/grpc"
	"github.com/bldsoft/geos/pkg/config"
	"github.com/bldsoft/geos/pkg/entity"
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

func geoNamesFilter(ctx *cli.Context) entity.GeoNameFilter {
	var countryCodes []string
	if codes := ctx.String("countries"); codes != "" {
		countryCodes = strings.Split(codes, ",")
	}
	inGeoNamesIDs := ctx.Uint64Slice("geoname-ids")
	outGeoNamesIDs := make([]uint32, 0, len(inGeoNamesIDs))
	for _, id := range inGeoNamesIDs {
		outGeoNamesIDs = append(outGeoNamesIDs, uint32(id))
	}
	return entity.GeoNameFilter{
		CountryCodes: countryCodes,
		NamePrefix:   ctx.String("name-prefix"),
		Limit:        uint32(ctx.Int64("limit")),
		GeoNameIDs:   outGeoNamesIDs,
	}
}

func commonFlags() []cli.Flag {
	var defaults config.Config
	_ = gost.SetDefaults(&defaults)
	return []cli.Flag{
		&cli.StringFlag{
			Name:        "host",
			Usage:       "Service host",
			Value:       defaults.Server.ServiceAddress.Host(),
			Destination: &host,
			Aliases:     []string{"H"},
		},
		&cli.UintFlag{
			Name:        "port",
			Usage:       "Service gRPC port",
			Value:       uint(defaults.GRPCServiceAddress.PortInt()),
			Destination: &port,
			Aliases:     []string{"p"},
		},
	}
}

func commonGeoNamesFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    "countries",
			Usage:   "Comma separated list of country codes",
			Aliases: []string{"cc"},
		},
		&cli.StringFlag{
			Name:    "name-prefix",
			Value:   "",
			Usage:   "Name prefix",
			Aliases: []string{"np"},
		},
		&cli.Int64Flag{
			Name:    "limit",
			Aliases: []string{"l"},
		},
		&cli.Uint64SliceFlag{
			Name:    "geoname-ids",
			Aliases: []string{"gid"},
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
					city, err := client(ctx).City(ctx.Context, addr(ctx), true)
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
				Name: "geoname-continent",
				Action: func(ctx *cli.Context) error {
					return print(client(ctx).GeoNameContinents(ctx.Context))
				},
			},
			{
				Name:  "geoname-country",
				Flags: commonGeoNamesFlags(),
				Action: func(ctx *cli.Context) error {
					countries, err := client(ctx).GeoNameCountries(ctx.Context, geoNamesFilter(ctx))
					if err != nil {
						return err
					}
					return print(countries)
				},
			},
			{
				Name:  "geoname-subdivision",
				Flags: commonGeoNamesFlags(),
				Action: func(ctx *cli.Context) error {
					subdivisions, err := client(ctx).GeoNameSubdivisions(ctx.Context, geoNamesFilter(ctx))
					if err != nil {
						return err
					}
					return print(subdivisions)
				},
			},
			{
				Name:  "geoname-city",
				Flags: commonGeoNamesFlags(),
				Action: func(ctx *cli.Context) error {

					cities, err := client(ctx).GeoNameCities(ctx.Context, geoNamesFilter(ctx))
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
