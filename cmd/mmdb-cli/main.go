package main

import (
	"fmt"
	"io"
	"log"
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
						source := source.NewMMDBSource(path, "")
						db, err := maxmind.Open(ctx.Context, source)
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
					source := source.NewMMDBSource(paths[0], "")
					db, err := maxmind.Open(ctx.Context, source)
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
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
