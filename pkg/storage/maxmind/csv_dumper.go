package maxmind

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/csv"
	"io"

	"github.com/bldsoft/gost/log"
	"github.com/oschwald/maxminddb-golang"
)

type MaxmindCSVDumper[T CSVEntity] struct {
	Database
}

func NewCSVDumper[T CSVEntity](db Database) *MaxmindCSVDumper[T] {
	return &MaxmindCSVDumper[T]{db}
}
func (db MaxmindCSVDumper[T]) WriteCSVTo(ctx context.Context, w io.Writer) error {
	networks, err := db.Networks(ctx, maxminddb.SkipAliasedNetworks)
	if err != nil {
		return err
	}

	meta, err := db.MetaData(ctx)
	if err != nil {
		return err
	}
	writtenRows := 0
	precent := 1 + int(meta.NodeCount)/100

	csvWriter := csv.NewWriter(w)
	writeRow := func(row []string) error {
		err := csvWriter.Write(row)
		writtenRows++
		if writtenRows%(precent) == 0 {
			percents := writtenRows / (precent)
			log.FromContext(ctx).Debugf("CSV writing: %d%%", percents)
		}
		return err
	}

	if !networks.Next() {
		return nil
	}
	var record T
	subnet, err := networks.Network(&record)
	if err != nil {
		return err
	}

	names, row, err := record.MarshalCSV()
	if err != nil {
		return err
	}

	header := make([]string, 0, len(names)+1)
	header = append(header, "network")
	header = append(header, names...)
	if err := csvWriter.Write(header); err != nil {
		return err
	}
	csvRow := header
	csvRow[0] = subnet.String()
	copy(csvRow[1:], row)
	if err := writeRow(csvRow); err != nil {
		return err
	}

	for networks.Next() {
		var record T
		subnet, err := networks.Network(&record)
		if err != nil {
			return err
		}
		_, row, err := record.MarshalCSV()
		if err != nil {
			return err
		}
		csvRow[0] = subnet.String()
		copy(csvRow[1:], row)

		err = writeRow(csvRow)
		if err != nil {
			return err
		}
	}

	csvWriter.Flush()

	log.FromContext(ctx).Debugf("CSV writing: 100%%")
	return nil
}

func (db MaxmindCSVDumper[T]) CSV(ctx context.Context, gzipCompress bool) (io.Reader, error) {
	var buf bytes.Buffer
	var w io.Writer = &buf
	if gzipCompress {
		w = gzip.NewWriter(w)
	}
	db.WriteCSVTo(ctx, w)
	return &buf, nil
}
