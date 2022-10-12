package entity

// type Country = geoip2.Country

type Country struct {
	Continent struct {
		Code      string            `maxminddb:"code" json:"code,omitempty"`
		GeoNameID uint              `maxminddb:"geoname_id" json:"geoNameID,omitempty"`
		Names     map[string]string `maxminddb:"names" json:"names,omitempty"`
	} `maxminddb:"continent" json:"continent,omitempty"`
	Country struct {
		GeoNameID         uint              `maxminddb:"geoname_id" json:"geoNameID,omitempty"`
		IsInEuropeanUnion bool              `maxminddb:"is_in_european_union" json:"isInEuropeanUnion,omitempty"`
		IsoCode           string            `maxminddb:"iso_code" json:"isoCode,omitempty"`
		Names             map[string]string `maxminddb:"names" json:"names,omitempty"`
	} `maxminddb:"country" json:"country,omitempty"`
	RegisteredCountry struct {
		GeoNameID         uint              `maxminddb:"geoname_id" json:"geoNameID,omitempty"`
		IsInEuropeanUnion bool              `maxminddb:"is_in_european_union" json:"isInEuropeanUnion,omitempty"`
		IsoCode           string            `maxminddb:"iso_code" json:"isoCode,omitempty"`
		Names             map[string]string `maxminddb:"names" json:"names,omitempty"`
	} `maxminddb:"registered_country" json:"registeredCountry,omitempty"`
	RepresentedCountry struct {
		GeoNameID         uint              `maxminddb:"geoname_id" json:"geoNameID,omitempty"`
		IsInEuropeanUnion bool              `maxminddb:"is_in_european_union" json:"isInEuropeanUnion,omitempty"`
		IsoCode           string            `maxminddb:"iso_code" json:"isoCode,omitempty"`
		Names             map[string]string `maxminddb:"names" json:"names,omitempty"`
		Type              string            `maxminddb:"type" json:"type,omitempty"`
	} `maxminddb:"represented_country" json:"representedCountry,omitempty"`
	Traits struct {
		IsAnonymousProxy    bool `maxminddb:"is_anonymous_proxy" json:"isAnonymousProxy,omitempty"`
		IsSatelliteProvider bool `maxminddb:"is_satellite_provider" json:"isSatelliteProvider,omitempty"`
	} `maxminddb:"traits" json:"traits,omitempty"`
}
