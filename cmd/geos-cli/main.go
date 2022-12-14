package main

import (
	"context"
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
	flags := commonFlags()

	app := &cli.App{
		Name:  "geos-cli",
		Usage: "Geos gRPC client",
		Commands: []*cli.Command{
			{
				Name:  "city",
				Flags: flags,
				Action: func(ctx *cli.Context) error {
					city, err := client(ctx).City(context.Background(), addr(ctx))
					if err != nil {
						return err
					}
					return print(city)
				},
			},
			{
				Name:  "country",
				Flags: flags,
				Action: func(ctx *cli.Context) error {
					country, err := client(ctx).Country(context.Background(), addr(ctx))
					if err != nil {
						return err
					}
					return print(country)
				},
			},
			{
				Name: "city-lite",
				Flags: append(flags,
					&cli.StringFlag{
						Name:    "lang",
						Value:   "en",
						Usage:   "Language for country 1and city name",
						Aliases: []string{"l"}},
				),
				Action: func(ctx *cli.Context) error {
					cityLite, err := client(ctx).CityLite(context.Background(), addr(ctx), ctx.String("lang"))
					if err != nil {
						return err
					}
					return print(cityLite)
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
