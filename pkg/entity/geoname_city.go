package entity

import (
	"encoding/json"

	"github.com/mkrou/geonames/models"
)

type GeoName struct {
	*models.Geoname
	ContinentCode   string `csv:"continent code"`
	ContinentName   string `csv:"continent name"`
	CountryName     string `csv:"country name"`
	SubdivisionName string `csv:"subdivision name"`
}

type geoNameJson struct {
	*modelGeoNameJson
	ContinentCode   string `csv:"continent code" json:"continentCode"`
	ContinentName   string `csv:"continent name" json:"continentName"`
	CountryName     string `csv:"country name" json:"countryName"`
	SubdivisionName string `csv:"subdivision name" json:"subdivisionName"`
}

// same as models.Geoname, but with json tags
type modelGeoNameJson struct {
	Id                    int         `csv:"geonameid" valid:"required" json:"geoNameID"`
	Name                  string      `csv:"name" valid:"required" json:"name"`
	AsciiName             string      `csv:"asciiname" json:"asciiName"`
	AlternateNames        string      `csv:"alternatenames" json:"alternateNames"`
	Latitude              float64     `csv:"latitude" json:"latitude"`
	Longitude             float64     `csv:"longitude" json:"longitude"`
	Class                 string      `csv:"feature class" json:"class"`
	Code                  string      `csv:"feature code" json:"code"`
	CountryCode           string      `csv:"country code" json:"countryCode"`
	AlternateCountryCodes string      `csv:"cc2" json:"alternateCountryCodes"`
	Admin1Code            string      `csv:"admin1 code" json:"admin1Code"`
	Admin2Code            string      `csv:"admin2 code" json:"admin2Code"`
	Admin3Code            string      `csv:"admin3 code" json:"admin3Code"`
	Admin4Code            string      `csv:"admin4 code" json:"admin4Code"`
	Population            int         `csv:"population" json:"population"`
	Elevation             int         `csv:"elevation,omitempty" json:"elevation,omitempty"`
	DigitalElevationModel int         `csv:"dem,omitempty" json:"dem,omitempty"`
	Timezone              string      `csv:"timezone" json:"timezone"`
	ModificationDate      models.Time `csv:"modification date" valid:"required" json:"-"`
}

func (g GeoName) GetGeoNameID() int {
	return g.Geoname.Id
}

func (g GeoName) GetName() string {
	return g.Geoname.Name
}

func (g GeoName) GetContinentCode() string {
	return g.ContinentCode
}

func (g GeoName) GetContinentName() string {
	return g.ContinentName
}

func (g GeoName) GetCountryCode() string {
	return g.Geoname.CountryCode
}

func (g GeoName) GetCountryName() string {
	return g.CountryName
}

func (g GeoName) GetSubdivisionName() string {
	return g.SubdivisionName
}

func (g GeoName) GetCityName() string {
	return g.Name
}

func (g GeoName) GetTimeZone() string {
	return g.Timezone
}

func (s GeoName) MarshalJSON() ([]byte, error) {
	return json.Marshal(&geoNameJson{
		modelGeoNameJson: (*modelGeoNameJson)(s.Geoname),
		ContinentCode:    s.ContinentCode,
		ContinentName:    s.ContinentName,
		CountryName:      s.CountryName,
		SubdivisionName:  s.SubdivisionName,
	})
}

func (s *GeoName) UnmarshalJSON(data []byte) error {
	var geoNameJson geoNameJson
	if err := json.Unmarshal(data, &geoNameJson); err != nil {
		return err
	}
	s.Geoname = (*models.Geoname)(geoNameJson.modelGeoNameJson)
	s.ContinentCode = geoNameJson.ContinentCode
	s.ContinentName = geoNameJson.ContinentName
	s.CountryName = geoNameJson.CountryName
	s.SubdivisionName = geoNameJson.SubdivisionName
	return nil
}
