package grpc

import (
	pb "github.com/bldsoft/geos/pkg/controller/grpc/proto"
	"github.com/bldsoft/geos/pkg/entity"
)

func GeoNameCountryToPb(c *entity.GeoNameCountry) *pb.GeoNameCountryResponse {
	return &pb.GeoNameCountryResponse{
		IsoCode:            c.Iso2Code,
		Iso3Code:           c.Iso3Code,
		IsoNumeric:         c.IsoNumeric,
		Fips:               c.Fips,
		Name:               c.Name,
		Capital:            c.Capital,
		Area:               c.Area,
		Population:         int64(c.Population),
		Continent:          c.Continent,
		Tld:                c.Tld,
		CurrencyCode:       c.CurrencyCode,
		CurrencyName:       c.CurrencyName,
		Phone:              c.Phone,
		PostalCodeFormat:   c.PostalCodeFormat,
		PostalCodeRegex:    c.PostalCodeRegex,
		Languages:          c.Languages,
		GeoNameId:          uint32(c.GeonameID),
		Neighbours:         c.Neighbours,
		EquivalentFipsCode: c.EquivalentFipsCode,
	}
}

func GeoNameSubdivisionToPb(s *entity.AdminSubdivision) *pb.GeoNameSubdivisionResponse {
	return &pb.GeoNameSubdivisionResponse{
		Code:      s.Code,
		Name:      s.Name,
		AsciiName: s.AsciiName,
		GeoNameId: uint32(s.GeonameId),
	}
}

func GeoNameCityToPb(c *entity.Geoname) *pb.GeoNameCityResponse {
	return &pb.GeoNameCityResponse{
		GeoNameId:             uint32(c.Id),
		Name:                  c.Name,
		AsciiName:             c.AsciiName,
		Latitude:              c.Latitude,
		Longitude:             c.Longitude,
		Class:                 c.Class,
		Code:                  c.Code,
		CountryCode:           c.CountryCode,
		AlternateCountryCodes: c.AlternateCountryCodes,
		Admin1Code:            c.Admin1Code,
		Admin2Code:            c.Admin2Code,
		Admin3Code:            c.Admin3Code,
		Admin4Code:            c.Admin4Code,
		Population:            int64(c.Population),
		Elevation:             int64(c.Elevation),
		DigitalElevationModel: int64(c.DigitalElevationModel),
		TimeZone:              c.Timezone,
	}
}
