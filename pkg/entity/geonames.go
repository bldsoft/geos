package entity

import (
	"slices"
	"strings"
)

type GeoNameEntity interface {
	GetGeoNameID() int
	GetName() string //own name
	GetContinentCode() string
	GetContinentName() string
	GetCountryCode() string
	GetCountryName() string
	GetSubdivisionName() string
	GetCityName() string
	GetTimeZone() string
}

type GeoNameFilter struct {
	GeoNameIDs   []uint32 `schema:"geoname-ids" json:"geonameIds"`
	CountryCodes []string `schema:"country-codes" json:"countryCodes"`
	NamePrefix   string   `schema:"name-prefix" json:"namePrefix"`
	Limit        uint32   `schema:"limit" json:"limit"`
}

func (f *GeoNameFilter) Match(e GeoNameEntity) bool {
	if len(f.GeoNameIDs) > 0 && !slices.Contains(f.GeoNameIDs, uint32(e.GetGeoNameID())) {
		return false
	}

	if len(f.CountryCodes) > 0 && !slices.Contains(f.CountryCodes, e.GetCountryCode()) {
		return false
	}

	if len(f.NamePrefix) > 0 && !strings.HasPrefix(e.GetName(), f.NamePrefix) {
		return false
	}

	return true
}

var _ GeoNameEntity = &GeoNameCountry{}
var _ GeoNameEntity = &GeoNameAdminSubdivision{}
var _ GeoNameEntity = &GeoName{}
var _ GeoNameEntity = &GeoNameContinent{}
