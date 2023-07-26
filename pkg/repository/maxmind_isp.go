package repository

import (
	"bytes"
	"context"
	"encoding/csv"
	"io"
	"strconv"

	"github.com/bldsoft/geos/pkg/entity"
	"github.com/oschwald/maxminddb-golang"
)

type ispDB struct {
	*database
}

func openISPDB(path string, required bool) *ispDB {
	return &ispDB{openDB(path, MaxmindDBTypeISP, required)}
}

func (db *ispDB) WriteCSVTo(ctx context.Context, w io.Writer) error {
	networks, err := db.Networks(maxminddb.SkipAliasedNetworks)
	if err != nil {
		return err
	}

	csvWriter := csv.NewWriter(w)

	if err := csvWriter.Write([]string{
		"network",
		"autonomous_system_organization",
		"ISP",
		"mobile_country_code",
		"mobile_network_code",
		"organization",
		"autonomous_system_number",
	}); err != nil {
		return err
	}

	var record entity.ISP
	for networks.Next() {
		subnet, err := networks.Network(&record)
		if err != nil {
			return err
		}

		err = csvWriter.Write([]string{
			subnet.String(),
			record.AutonomousSystemOrganization,
			record.ISP,
			record.MobileCountryCode,
			record.MobileCountryCode,
			record.Organization,
			strconv.FormatUint(uint64(record.AutonomousSystemNumber), 10),
		})
		if err != nil {
			return err
		}
	}

	csvWriter.Flush()
	return nil
}

func (db *ispDB) CSV(ctx context.Context, withColumnNames bool) ([]byte, error) {
	var buf bytes.Buffer
	db.WriteCSVTo(ctx, &buf)
	return buf.Bytes(), nil
}
