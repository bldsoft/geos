package entity

type CityLite struct {
	City struct {
		Name string
	}
	Country struct {
		ISOCode string
		Name    string
	}
	Location LocationLite
}
type CityLiteDb struct {
	City struct {
		Names map[string]string `maxminddb:"names"`
	} `maxminddb:"city"`
	Country struct {
		ISOCode string            `maxminddb:"iso_code"`
		Names   map[string]string `maxminddb:"names"`
	} `maxminddb:"country"`
	Location LocationLite `maxminddb:"location"`
}

type LocationLite struct {
	Latitude  float64 `maxminddb:"latitude"`
	Longitude float64 `maxminddb:"longitude"`
	TimeZone  string  `maxminddb:"time_zone"`
}

func DbToCityLite(cityLiteDb *CityLiteDb, lang string) *CityLite {
	var cityLite CityLite
	cityLite.City.Name = cityLiteDb.City.Names[lang]
	cityLite.Country.ISOCode = cityLiteDb.Country.ISOCode
	cityLite.Country.Name = cityLiteDb.Country.Names[lang]
	cityLite.Location = cityLiteDb.Location
	return &cityLite
}
