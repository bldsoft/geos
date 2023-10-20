package repository

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/csv"
	"io"
	"strconv"

	"github.com/bldsoft/geos/pkg/entity"
	"github.com/oschwald/maxminddb-golang"
)

type cityDB struct {
	*database
}

func openCityDB(path string, required bool) *cityDB {
	return &cityDB{openDB(path, MaxmindDBTypeCity, required)}
}

func (db *cityDB) WriteCSVTo(ctx context.Context, w io.Writer) error {
	networks, err := db.Networks(maxminddb.SkipAliasedNetworks)
	if err != nil {
		return err
	}

	csvWriter := csv.NewWriter(w)

	if err := csvWriter.Write([]string{
		"network",
		"city_geoname_id",
		"subdivision_geoname_id",
		"registered_country_geoname_id",
		"represented_country_geoname_id",
		"is_anonymous_proxy",
		"is_satellite_provider",
		"latitude",
		"longitude",
		"accuracy_radius",
	}); err != nil {
		return err
	}

	var record entity.City
	for networks.Next() {
		subnet, err := networks.Network(&record)
		if err != nil {
			return err
		}

		var subdivisionGeonameID uint64
		if len(record.Subdivisions) > 0 {
			subdivisionGeonameID = uint64(record.Subdivisions[0].GeoNameID)
		}
		err = csvWriter.Write([]string{
			subnet.String(),
			strconv.FormatUint(uint64(record.City.GeoNameID), 10),
			strconv.FormatUint(subdivisionGeonameID, 10),
			strconv.FormatUint(uint64(record.RegisteredCountry.GeoNameID), 10),
			strconv.FormatUint(uint64(record.RepresentedCountry.GeoNameID), 10),
			formatBool(record.Traits.IsAnonymousProxy),
			formatBool(record.Traits.IsSatelliteProvider),
			strconv.FormatFloat(record.Location.Latitude, 'f', 4, 64),
			strconv.FormatFloat(record.Location.Longitude, 'f', 4, 64),
			strconv.FormatUint(uint64(record.Location.AccuracyRadius), 10),
		})
		if err != nil {
			return err
		}
	}

	csvWriter.Flush()
	return nil
}

func (db *cityDB) CSV(ctx context.Context, gzipCompress bool) (io.Reader, error) {
	var buf bytes.Buffer
	var w io.Writer = &buf
	if gzipCompress {
		w = gzip.NewWriter(w)
	}
	db.WriteCSVTo(ctx, w)
	return &buf, nil
}

func formatBool(b bool) string {
	if b {
		return "1"
	}
	return "0"
}
