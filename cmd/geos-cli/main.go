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
)

func client(ctx *cli.Context) *grpc.Client {
	var defaults config.Config
	defaults.SetDefaults()
	if port := ctx.Int("port"); port != 0 {
		defaults.GrpcPort = port
	}
	if host := ctx.String("host"); len(host) != 0 {
		defaults.Server.Host = host
	}
	client, err := grpc.NewClient(defaults.GrpcAddr())
	if err != nil {
		log.Fatalf(err.Error())
	}
	return client
}

func ip(ctx *cli.Context) string {
	ip := ctx.Args().Get(0)
	if len(ip) == 0 {
		return "me"
	}
	return ip
}

func main() {
	app := &cli.App{
		Name:  "Geos grpc test client",
		Usage: "geos-cli",
		Commands: []*cli.Command{
			{
				Name: "city",
				Action: func(ctx *cli.Context) error {
					city, err := client(ctx).City(context.Background(), ip(ctx))
					if err != nil {
						return err
					}
					data, err := json.MarshalIndent(city, "", "	")
					if err != nil {
						return err
					}
					fmt.Printf("%s", data)
					return nil
				},
			},
			{
				Name: "country",
				Action: func(ctx *cli.Context) error {
					country, err := client(ctx).Country(context.Background(), ip(ctx))
					if err != nil {
						return err
					}
					data, err := json.MarshalIndent(country, "", "	")
					if err != nil {
						return err
					}
					fmt.Printf("%s", data)
					return nil
				},
			},
			{
				Name: "city-lite",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "lang",
						Aliases: []string{"l"}},
				},
				Action: func(ctx *cli.Context) error {
					cityLite, err := client(ctx).CityLite(context.Background(), ip(ctx), ctx.String("lang"))
					if err != nil {
						return err
					}
					data, err := json.MarshalIndent(cityLite, "", "	")
					if err != nil {
						return err
					}
					fmt.Printf("%s", data)
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
