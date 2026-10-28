package maxmind

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/csv"
	"io"

	"github.com/bldsoft/gost/log"
)

type MaxmindCSVDumper[T CSVEntity] struct {
	Database
}

func NewCSVDumper[T CSVEntity](db Database) *MaxmindCSVDumper[T] {
	return &MaxmindCSVDumper[T]{db}
}

func (db MaxmindCSVDumper[T]) WriteCSVTo(ctx context.Context, w io.Writer) error {
	networks, err := db.Networks(ctx)
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

	var csvRow []string
	first := true
	for result := range networks {
		if err := result.Err(); err != nil {
			return err
		}
		subnet := result.Prefix()
		var record T
		if err := result.Decode(&record); err != nil {
			return err
		}
		names, row, err := record.MarshalCSV()
		if err != nil {
			return err
		}
		if first {
			header := make([]string, 0, len(names)+1)
			header = append(header, "network")
			header = append(header, names...)
			if err := csvWriter.Write(header); err != nil {
				return err
			}
			csvRow = make([]string, len(header))
			first = false
		}
		csvRow[0] = subnet.String()
		copy(csvRow[1:], row)
		if err := writeRow(csvRow); err != nil {
			return err
		}
	}

	csvWriter.Flush()

	log.FromContext(ctx).Debugf("CSV writing: 100%%")
	return nil
}

func (db MaxmindCSVDumper[T]) CSV(ctx context.Context, gzipCompress bool) (io.Reader, error) {
	var buf bytes.Buffer
	if gzipCompress {
		gz := gzip.NewWriter(&buf)
		if err := db.WriteCSVTo(ctx, gz); err != nil {
			return nil, err
		}
		if err := gz.Close(); err != nil {
			return nil, err
		}
		return &buf, nil
	}
	if err := db.WriteCSVTo(ctx, &buf); err != nil {
		return nil, err
	}
	return &buf, nil
}
