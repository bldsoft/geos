package entity

type CityLite struct {
	City struct {
		Name string `json:"name,omitempty"`
	} `json:"city"`
	Country struct {
		ISOCode string `json:"isoCode,omitempty"`
		Name    string `json:"name,omitempty"`
	} `json:"country"`
	Location LocationLite `json:"location"`
}
type CityLiteDb struct {
	City struct {
		Names map[string]string `maxminddb:"names" json:"names,omitempty"`
	} `maxminddb:"city" json:"city"`
	Country struct {
		ISOCode string            `maxminddb:"iso_code" json:"isoCode,omitempty"`
		Names   map[string]string `maxminddb:"names" json:"names,omitempty"`
	} `maxminddb:"country" json:"country"`
	Location LocationLite `maxminddb:"location" json:"location"`
}

type LocationLite struct {
	Latitude  float64 `maxminddb:"latitude" json:"latitude,omitempty"`
	Longitude float64 `maxminddb:"longitude" json:"longitude,omitempty"`
	TimeZone  string  `maxminddb:"time_zone" json:"timeZone,omitempty"`
}

func DbToCityLite(cityLiteDb *CityLiteDb, lang string) *CityLite {
	var cityLite CityLite
	cityLite.City.Name = cityLiteDb.City.Names[lang]
	cityLite.Country.ISOCode = cityLiteDb.Country.ISOCode
	cityLite.Country.Name = cityLiteDb.Country.Names[lang]
	cityLite.Location = cityLiteDb.Location
	return &cityLite
}
