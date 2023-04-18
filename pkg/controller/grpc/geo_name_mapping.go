package grpc

import (
	pb "github.com/bldsoft/geos/pkg/controller/grpc/proto"
	"github.com/bldsoft/geos/pkg/entity"
)

func GeoNameContinentToPb(c *entity.GeoNameContinent) *pb.GeoNameContinentResponse {
	return &pb.GeoNameContinentResponse{
		Code:      c.Code,
		Name:      c.Name,
		GeoNameId: uint32(c.GeonameID),
	}
}

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

func PbToGeoNameContinent(c *pb.GeoNameContinentResponse) *entity.GeoNameContinent {
	return &entity.GeoNameContinent{
		Code:      c.Code,
		Name:      c.Name,
		GeonameID: int(c.GeoNameId),
	}
}

func PbToGeoNameCountry(c *pb.GeoNameCountryResponse) *entity.GeoNameCountry {
	return &entity.GeoNameCountry{
		Iso2Code:           c.IsoCode,
		Iso3Code:           c.Iso3Code,
		IsoNumeric:         c.IsoNumeric,
		Fips:               c.Fips,
		Name:               c.Name,
		Capital:            c.Capital,
		Area:               c.Area,
		Population:         int(c.Population),
		Continent:          c.Continent,
		Tld:                c.Tld,
		CurrencyCode:       c.CurrencyCode,
		CurrencyName:       c.CurrencyName,
		Phone:              c.Phone,
		PostalCodeFormat:   c.PostalCodeFormat,
		PostalCodeRegex:    c.PostalCodeRegex,
		Languages:          c.Languages,
		GeonameID:          int(c.GeoNameId),
		Neighbours:         c.Neighbours,
		EquivalentFipsCode: c.EquivalentFipsCode,
	}
}

func PbToGeoNameSubdivision(s *pb.GeoNameSubdivisionResponse) *entity.AdminSubdivision {
	return &entity.AdminSubdivision{
		Code:      s.Code,
		Name:      s.Name,
		AsciiName: s.AsciiName,
		GeonameId: int(s.GeoNameId),
	}
}

func PbToGeoNameCity(c *pb.GeoNameCityResponse) *entity.Geoname {
	return &entity.Geoname{
		Id:                    int(c.GeoNameId),
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
		Population:            int(c.Population),
		Elevation:             int(c.Elevation),
		DigitalElevationModel: int(c.DigitalElevationModel),
		Timezone:              c.TimeZone,
	}
}

func PbGeoNameRequestToFilter(r *pb.GeoNameRequest) entity.GeoNameFilter {
	return entity.GeoNameFilter{
		CountryCodes: r.CountryCodes,
		Search:       r.Search,
	}
}

func FilterToPbGeoNameRequest(f entity.GeoNameFilter) *pb.GeoNameRequest {
	return &pb.GeoNameRequest{
		CountryCodes: f.CountryCodes,
		Search:       f.Search,
	}
}
