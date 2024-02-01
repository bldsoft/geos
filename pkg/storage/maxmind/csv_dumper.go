package maxmind

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/csv"
	"io"

	"github.com/oschwald/maxminddb-golang"
)

type MaxmindCSVDumper[T CSVEntity] struct {
	Database
}

func NewCSVDumper[T CSVEntity](db Database) *MaxmindCSVDumper[T] {
	return &MaxmindCSVDumper[T]{db}
}

func (db MaxmindCSVDumper[T]) WriteCSVTo(ctx context.Context, w io.Writer) error {
	networks, err := db.Networks(maxminddb.SkipAliasedNetworks)
	if err != nil {
		return err
	}

	csvWriter := csv.NewWriter(w)

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
	if err := csvWriter.Write(csvRow); err != nil {
		return err
	}

	for networks.Next() {
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

		err = csvWriter.Write(csvRow)
		if err != nil {
			return err
		}
	}

	csvWriter.Flush()
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
