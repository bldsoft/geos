package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"

	"github.com/bldsoft/geos/pkg/storage/maxmind"
	"github.com/bldsoft/geos/pkg/storage/source"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "mmdb-cli",
		Usage: "MMDB CLI tool",
		Commands: []*cli.Command{
			{
				Name: "merge",
				Action: func(ctx *cli.Context) (err error) {
					paths := ctx.Args().Slice()
					if len(paths) < 2 {
						return fmt.Errorf("at least two paths are required")
					}
					outFile, err := os.OpenFile(paths[len(paths)-1], os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
					if err != nil {
						return err
					}

					defer func() {
						if err != nil {
							os.Remove(paths[len(paths)-1])
						}
					}()
					defer outFile.Close()

					res := maxmind.NewMultiMaxMindDB()
					for _, path := range paths[:len(paths)-1] {
						db, err := openDatabase(ctx.Context, path)
						if err != nil {
							return err
						}
						res = res.Add(db)
					}

					reader, err := res.RawData(ctx.Context)
					if err != nil {
						return err
					}

					_, err = io.Copy(outFile, reader)
					return err
				},
			},
			{
				Name: "metadata",
				Action: func(ctx *cli.Context) error {
					paths := ctx.Args().Slice()
					db, err := openDatabase(ctx.Context, paths[0])
					if err != nil {
						return err
					}
					meta, err := db.MetaData(ctx.Context)
					if err != nil {
						return err
					}
					fmt.Printf("Database type: %s\n", meta.DatabaseType)
					fmt.Printf("Binary format major version: %d\n", meta.BinaryFormatMajorVersion)
					fmt.Printf("Binary format minor version: %d\n", meta.BinaryFormatMinorVersion)
					fmt.Printf("Build epoch: %d (%s)\n", meta.BuildEpoch, time.Unix(int64(meta.BuildEpoch), 0).Format(time.RFC3339))
					fmt.Printf("IP version: %d\n", meta.IPVersion)
					fmt.Printf("Node count: %d\n", meta.NodeCount)
					fmt.Printf("Record size: %d\n", meta.RecordSize)
					fmt.Printf("Languages: %v\n", meta.Languages)
					fmt.Printf("Description: %v\n", meta.Description)
					return nil
				},
			},
			{
				Name: "lookup",
				Action: func(ctx *cli.Context) error {
					dbPath := ctx.Args().Get(0)
					ip := ctx.Args().Get(1)
					if ip == "" {
						return fmt.Errorf("ip is required")
					}
					db, err := openDatabase(ctx.Context, dbPath)
					if err != nil {
						return err
					}
					var res map[string]interface{}
					err = db.Lookup(ctx.Context, net.ParseIP(ip), &res)
					if err != nil {
						return err
					}
					if len(res) == 0 {
						return fmt.Errorf("no result found")
					}
					for k, v := range res {
						fmt.Printf("%s: %v\n", k, v)
					}
					return nil
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func openDatabase(ctx context.Context, dbPath string) (*maxmind.MaxmindDatabase, error) {
	source := source.NewMMDBSource(dbPath, "")
	db, err := maxmind.Open(ctx, source)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %s", dbPath)
	}
	return db, nil
}
