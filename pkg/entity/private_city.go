package entity

var PrivateCity City

func init() {
	const PrivateGeoNameID = 9999999
	PrivateCity.City.GeoNameID = PrivateGeoNameID
	PrivateCity.City.Names = map[string]string{"en": "Private network"}

	PrivateCity.Country.GeoNameID = PrivateGeoNameID
	PrivateCity.Country.Names = map[string]string{"en": "Private network"}

	PrivateCity.Subdivisions = append(PrivateCity.Subdivisions, struct {
		GeoNameID uint              `maxminddb:"geoname_id" json:"geoNameID,omitempty"`
		IsoCode   string            `maxminddb:"iso_code" json:"isoCode,omitempty"`
		Names     map[string]string `maxminddb:"names" json:"names,omitempty"`
	}{
		GeoNameID: PrivateGeoNameID,
		IsoCode:   "",
		Names:     map[string]string{"en": "Private network"},
	})

	PrivateCity.Continent.GeoNameID = PrivateGeoNameID
	PrivateCity.Continent.Names = map[string]string{"en": "Private network"}
}
