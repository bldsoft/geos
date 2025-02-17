package grpc

import (
	pb "github.com/bldsoft/geos/pkg/controller/grpc/proto"
	"github.com/bldsoft/geos/pkg/entity"
	"github.com/mkrou/geonames/models"
)

func GeoNameContinentToPb(c *entity.GeoNameContinent) *pb.GeoNameContinentResponse {
	return &pb.GeoNameContinentResponse{
		Code:      c.Code(),
		Name:      c.GetName(),
		GeoNameId: uint32(c.GetGeoNameID()),
	}
}

func GeoNameCountryToPb(c *entity.GeoNameCountry) *pb.GeoNameCountryResponse {
	return &pb.GeoNameCountryResponse{
		IsoCode:            c.Iso2Code,
		Iso3Code:           c.Iso3Code,
		IsoNumeric:         c.IsoNumeric,
		Fips:               c.Fips,
		Name:               c.GetName(),
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

func GeoNameSubdivisionToPb(s *entity.GeoNameAdminSubdivision) *pb.GeoNameSubdivisionResponse {
	return &pb.GeoNameSubdivisionResponse{
		Code:      s.Code,
		Name:      s.GetName(),
		AsciiName: s.AsciiName,
		GeoNameId: uint32(s.GeonameId),
	}
}

func GeoNameCityToPb(c *entity.GeoName) *pb.GeoNameCityResponse {
	return &pb.GeoNameCityResponse{
		GeoNameId:             uint32(c.Id),
		Name:                  c.GetName(),
		AsciiName:             c.AsciiName,
		Latitude:              c.Latitude,
		Longitude:             c.Longitude,
		Class:                 c.Class,
		Code:                  c.Code,
		CountryCode:           c.GetCountryCode(),
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
	return entity.NewGeoNameContinent(int(c.GeoNameId), c.Name, c.Code)
}

func PbToGeoNameCountry(c *pb.GeoNameCountryResponse) *entity.GeoNameCountry {
	return &entity.GeoNameCountry{
		Country: &models.Country{
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
		},
	}
}

func PbToGeoNameSubdivision(s *pb.GeoNameSubdivisionResponse) *entity.GeoNameAdminSubdivision {
	return &entity.GeoNameAdminSubdivision{
		AdminDivision: &models.AdminDivision{
			Code:      s.Code,
			Name:      s.Name,
			AsciiName: s.AsciiName,
			GeonameId: int(s.GeoNameId),
		},
	}
}

func PbToGeoNameCity(c *pb.GeoNameCityResponse) *entity.GeoName {
	return &entity.GeoName{
		Geoname: &models.Geoname{
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
		},
	}
}

func PbGeoNameRequestToFilter(r *pb.GeoNameRequest) entity.GeoNameFilter {
	return entity.GeoNameFilter{
		CountryCodes: r.CountryCodes,
		NamePrefix:   r.NamePrefix,
		GeoNameIDs:   r.GeoNameIds,
		Limit:        r.Limit,
	}
}

func FilterToPbGeoNameRequest(f entity.GeoNameFilter) *pb.GeoNameRequest {
	return &pb.GeoNameRequest{
		CountryCodes: f.CountryCodes,
		NamePrefix:   f.NamePrefix,
		GeoNameIds:   f.GeoNameIDs,
		Limit:        f.Limit,
	}
}
