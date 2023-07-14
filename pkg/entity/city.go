package entity

// type City = geoip2.City

type City struct {
	City struct {
		GeoNameID uint              `maxminddb:"geoname_id" json:"geoNameID,omitempty"`
		Names     map[string]string `maxminddb:"names" json:"names,omitempty"`
	} `maxminddb:"city" json:"city,omitempty"`
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
	Location struct {
		AccuracyRadius uint16  `maxminddb:"accuracy_radius" json:"accuracyRadius,omitempty"`
		Latitude       float64 `maxminddb:"latitude" json:"latitude,omitempty"`
		Longitude      float64 `maxminddb:"longitude" json:"longitude,omitempty"`
		MetroCode      uint    `maxminddb:"metro_code" json:"metroCode,omitempty"`
		TimeZone       string  `maxminddb:"time_zone" json:"timeZone,omitempty"`
	} `maxminddb:"location" json:"location,omitempty"`
	Postal struct {
		Code string `maxminddb:"code" json:"code,omitempty"`
	} `maxminddb:"postal" json:"postal,omitempty"`
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
	Subdivisions []struct {
		GeoNameID uint              `maxminddb:"geoname_id" json:"geoNameID,omitempty"`
		IsoCode   string            `maxminddb:"iso_code" json:"isoCode,omitempty"`
		Names     map[string]string `maxminddb:"names" json:"names,omitempty"`
	} `maxminddb:"subdivisions" json:"subdivisions,omitempty"`
	Traits struct {
		IsAnonymousProxy    bool `maxminddb:"is_anonymous_proxy" json:"isAnonymousProxy,omitempty"`
		IsSatelliteProvider bool `maxminddb:"is_satellite_provider" json:"isSatelliteProvider,omitempty"`
	} `maxminddb:"traits" json:"traits,omitempty"`

	ISP *ISP `json:"ISP,omitempty"`
}

// The ISP struct corresponds to the data in the GeoIP2 ISP database.
type ISP struct {
	AutonomousSystemOrganization string `maxminddb:"autonomous_system_organization"`
	ISP                          string `maxminddb:"isp"`
	MobileCountryCode            string `maxminddb:"mobile_country_code"`
	MobileNetworkCode            string `maxminddb:"mobile_network_code"`
	Organization                 string `maxminddb:"organization"`
	AutonomousSystemNumber       uint   `maxminddb:"autonomous_system_number"`
}
