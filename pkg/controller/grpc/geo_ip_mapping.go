package grpc

import (
	pb "github.com/bldsoft/geos/pkg/controller/grpc/proto"
	"github.com/bldsoft/geos/pkg/entity"
)

func PbToCountry(countryPb *pb.CountryResponse) *entity.Country {
	var country entity.Country
	country.Continent.Code = countryPb.Continent.Code
	country.Continent.GeoNameID = uint(countryPb.Continent.GeoNameId)
	country.Continent.Names = countryPb.Continent.Names

	country.Country.GeoNameID = uint(countryPb.Country.GeoNameId)
	country.Country.IsInEuropeanUnion = countryPb.Country.IsInEuropeanUnion
	country.Country.IsoCode = countryPb.Country.IsoCode
	country.Country.Names = countryPb.Country.Names

	country.RegisteredCountry.GeoNameID = uint(countryPb.RegisteredCountry.GeoNameId)
	country.RegisteredCountry.IsInEuropeanUnion = countryPb.RegisteredCountry.IsInEuropeanUnion
	country.RegisteredCountry.IsoCode = countryPb.RegisteredCountry.IsoCode
	country.RegisteredCountry.Names = countryPb.RegisteredCountry.Names

	country.RepresentedCountry.GeoNameID = uint(countryPb.RepresentedCountry.GeoNameId)
	country.RepresentedCountry.IsInEuropeanUnion = countryPb.RepresentedCountry.IsInEuropeanUnion
	country.RepresentedCountry.IsoCode = countryPb.RepresentedCountry.IsoCode
	country.RepresentedCountry.Names = countryPb.RepresentedCountry.Names
	country.RepresentedCountry.Type = countryPb.RepresentedCountry.Type

	country.Traits.IsAnonymousProxy = countryPb.Traits.IsAnonymousProxy
	country.Traits.IsSatelliteProvider = countryPb.Traits.IsSatelliteProvider

	return &country
}

func CountryToPb(country *entity.Country) *pb.CountryResponse {
	return &pb.CountryResponse{
		Continent: &pb.Continent{
			Code:      country.Continent.Code,
			GeoNameId: uint32(country.Continent.GeoNameID),
			Names:     country.Continent.Names,
		},
		Country: &pb.Country{
			GeoNameId:         uint32(country.Country.GeoNameID),
			IsInEuropeanUnion: country.Country.IsInEuropeanUnion,
			IsoCode:           country.Country.IsoCode,
			Names:             country.Country.Names,
		},
		RegisteredCountry: &pb.Country{
			GeoNameId:         uint32(country.RegisteredCountry.GeoNameID),
			IsInEuropeanUnion: country.RegisteredCountry.IsInEuropeanUnion,
			IsoCode:           country.RegisteredCountry.IsoCode,
			Names:             country.RegisteredCountry.Names,
		},
		RepresentedCountry: &pb.RepresentedCountry{
			GeoNameId:         uint32(country.RepresentedCountry.GeoNameID),
			IsInEuropeanUnion: country.RepresentedCountry.IsInEuropeanUnion,
			IsoCode:           country.RepresentedCountry.IsoCode,
			Names:             country.RepresentedCountry.Names,
			Type:              country.RepresentedCountry.Type,
		},
		Traits: &pb.Traits{
			IsAnonymousProxy:    country.Traits.IsAnonymousProxy,
			IsSatelliteProvider: country.Traits.IsSatelliteProvider,
		},
	}
}

func PbToCity(cityPb *pb.CityResponse) *entity.City {
	var city entity.City

	city.City.GeoNameID = uint(cityPb.City.GeoNameId)
	city.City.Names = cityPb.City.Names

	city.Continent.Code = cityPb.Continent.Code
	city.Continent.GeoNameID = uint(cityPb.Continent.GeoNameId)
	city.Continent.Names = cityPb.Continent.Names

	city.Country.GeoNameID = uint(cityPb.Country.GeoNameId)
	city.Country.IsInEuropeanUnion = cityPb.Country.IsInEuropeanUnion
	city.Country.IsoCode = cityPb.Country.IsoCode
	city.Country.Names = cityPb.Country.Names

	city.Location.AccuracyRadius = uint16(cityPb.Location.AccuracyRadius)
	city.Location.Latitude = cityPb.Location.Latitude
	city.Location.Longitude = cityPb.Location.Longitude
	city.Location.MetroCode = uint(cityPb.Location.MetroCode)
	city.Location.TimeZone = cityPb.Location.TimeZone

	city.Postal.Code = cityPb.Postal.Code

	city.RegisteredCountry.GeoNameID = uint(cityPb.RegisteredCountry.GeoNameId)
	city.RegisteredCountry.IsInEuropeanUnion = cityPb.RegisteredCountry.IsInEuropeanUnion
	city.RegisteredCountry.IsoCode = cityPb.RegisteredCountry.IsoCode
	city.RegisteredCountry.Names = cityPb.RegisteredCountry.Names

	city.RepresentedCountry.GeoNameID = uint(cityPb.RepresentedCountry.GeoNameId)
	city.RepresentedCountry.IsInEuropeanUnion = cityPb.RepresentedCountry.IsInEuropeanUnion
	city.RepresentedCountry.IsoCode = cityPb.RepresentedCountry.IsoCode
	city.RepresentedCountry.Names = cityPb.RepresentedCountry.Names
	city.RepresentedCountry.Type = cityPb.RepresentedCountry.Type

	for _, subdivision := range cityPb.Subdivisions {
		city.Subdivisions = append(city.Subdivisions, struct {
			GeoNameID uint              `maxminddb:"geoname_id" json:"geoNameID,omitempty"`
			IsoCode   string            `maxminddb:"iso_code" json:"isoCode,omitempty"`
			Names     map[string]string `maxminddb:"names" json:"names,omitempty"`
		}{
			GeoNameID: uint(subdivision.GeoNameId),
			IsoCode:   subdivision.IsoCode,
			Names:     subdivision.Names,
		})
	}

	city.Traits.IsAnonymousProxy = cityPb.Traits.IsAnonymousProxy
	city.Traits.IsSatelliteProvider = cityPb.Traits.IsSatelliteProvider

	city.ISP = PbToISP(cityPb.Isp)
	return &city
}

func CityToPb(city *entity.City) *pb.CityResponse {
	subdivisions := make([]*pb.Subdivision, 0, len(city.Subdivisions))
	for _, subdivision := range city.Subdivisions {
		subdivisions = append(subdivisions, &pb.Subdivision{
			GeoNameId: uint32(subdivision.GeoNameID),
			IsoCode:   subdivision.IsoCode,
			Names:     subdivision.Names,
		})
	}
	return &pb.CityResponse{
		City: &pb.City{
			GeoNameId: uint32(city.City.GeoNameID),
			Names:     city.City.Names,
		},
		Continent: &pb.Continent{
			Code:      city.Continent.Code,
			GeoNameId: uint32(city.Continent.GeoNameID),
			Names:     city.Continent.Names,
		},
		Country: &pb.Country{
			GeoNameId:         uint32(city.Country.GeoNameID),
			IsInEuropeanUnion: city.Country.IsInEuropeanUnion,
			IsoCode:           city.Country.IsoCode,
			Names:             city.Country.Names,
		},
		Location: &pb.Location{
			AccuracyRadius: uint32(city.Location.AccuracyRadius),
			Latitude:       city.Location.Latitude,
			Longitude:      city.Location.Longitude,
			MetroCode:      uint32(city.Location.MetroCode),
			TimeZone:       city.Location.TimeZone,
		},
		Postal: &pb.Postal{
			Code: city.Postal.Code,
		},
		RegisteredCountry: &pb.Country{
			GeoNameId:         uint32(city.RegisteredCountry.GeoNameID),
			IsInEuropeanUnion: city.RegisteredCountry.IsInEuropeanUnion,
			IsoCode:           city.RegisteredCountry.IsoCode,
			Names:             city.RegisteredCountry.Names,
		},
		RepresentedCountry: &pb.RepresentedCountry{
			GeoNameId:         uint32(city.RepresentedCountry.GeoNameID),
			IsInEuropeanUnion: city.RepresentedCountry.IsInEuropeanUnion,
			IsoCode:           city.RepresentedCountry.IsoCode,
			Names:             city.RepresentedCountry.Names,
			Type:              city.RepresentedCountry.Type,
		},
		Subdivisions: subdivisions,
		Traits: &pb.Traits{
			IsAnonymousProxy:    city.Traits.IsAnonymousProxy,
			IsSatelliteProvider: city.Traits.IsSatelliteProvider,
		},
		Isp: ISPToPb(city.ISP),
	}
}

func CityLiteToPb(cityLite *entity.CityLite) *pb.CityLiteResponse {
	return &pb.CityLiteResponse{
		City: &pb.CityLiteResponse_City{
			Name: cityLite.City.Name,
		},
		Country: &pb.CityLiteResponse_Country{
			IsoCode: cityLite.Country.ISOCode,
			Name:    cityLite.Country.Name,
		},
		Location: &pb.CityLiteResponse_Location{
			Latitude:  cityLite.Location.Latitude,
			Longitude: cityLite.Location.Longitude,
			TimeZone:  cityLite.Location.TimeZone,
		},
	}
}

func PbToCityLite(cityLitePb *pb.CityLiteResponse) *entity.CityLite {
	var cityLite entity.CityLite
	cityLite.City.Name = cityLitePb.City.Name

	cityLite.Country.ISOCode = cityLitePb.Country.IsoCode
	cityLite.Country.Name = cityLitePb.Country.Name

	cityLite.Location.Latitude = cityLitePb.Location.Latitude
	cityLite.Location.Longitude = cityLitePb.Location.Longitude
	cityLite.Location.TimeZone = cityLitePb.Location.TimeZone

	return &cityLite
}

func ISPToPb(isp *entity.ISP) *pb.ISP {
	if isp == nil {
		return nil
	}
	return &pb.ISP{
		AutonomousSystemOrganization: isp.AutonomousSystemOrganization,
		Isp:                          isp.ISP,
		MobileCountryCode:            isp.MobileCountryCode,
		MobileNetworkCode:            isp.MobileNetworkCode,
		Organization:                 isp.Organization,
		AutonomousSystemNumber:       uint32(isp.AutonomousSystemNumber),
	}
}

func PbToISP(isp *pb.ISP) *entity.ISP {
	if isp == nil {
		return nil
	}
	return &entity.ISP{
		AutonomousSystemOrganization: isp.AutonomousSystemOrganization,
		ISP:                          isp.Isp,
		MobileCountryCode:            isp.MobileCountryCode,
		MobileNetworkCode:            isp.MobileNetworkCode,
		Organization:                 isp.Organization,
		AutonomousSystemNumber:       uint(isp.AutonomousSystemNumber),
	}
}
