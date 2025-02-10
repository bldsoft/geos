package entity

import (
	"encoding/json"

	"github.com/mkrou/geonames/models"
)

type GeoNameCountry struct {
	*models.Country
	ContinentName string
}

// same as models.Country, but with json tags
type geoNameCountryJson struct {
	Iso2Code           string  `csv:"ISO" valid:"required" json:"iso2Code"`
	Iso3Code           string  `csv:"ISO3" valid:"required" json:"iso3Code"`
	IsoNumeric         string  `csv:"ISO-Numeric" valid:"required" json:"isoNumeric"`
	Fips               string  `csv:"fips" json:"fips"`
	Name               string  `csv:"Country" valid:"required" json:"name"`
	Capital            string  `csv:"Capital" json:"capital"`
	Area               float64 `csv:"Area(in sq km)" json:"area"`
	Population         int     `csv:"Population" json:"population"`
	Continent          string  `csv:"Continent" valid:"required" json:"continent"`
	Tld                string  `csv:"tld" json:"tld"`
	CurrencyCode       string  `csv:"CurrencyCode" json:"currencyCode"`
	CurrencyName       string  `csv:"CurrencyName" json:"currencyName"`
	Phone              string  `csv:"Phone" json:"phone"`
	PostalCodeFormat   string  `csv:"Postal Code Format" json:"postalCodeFormat"`
	PostalCodeRegex    string  `csv:"Postal Code Regex" json:"postalCodeRegex"`
	Languages          string  `csv:"Languages" json:"languages"`
	GeonameID          int     `csv:"geonameid" valid:"required" json:"geoNameID"`
	Neighbours         string  `csv:"neighbours" json:"neighbours"`
	EquivalentFipsCode string  `csv:"EquivalentFipsCode" json:"equivalentFipsCode"`
}

func (s GeoNameCountry) GetGeoNameID() int {
	return s.Country.GeonameID
}

func (s GeoNameCountry) GetName() string {
	return s.Country.Name
}

func (s GeoNameCountry) GetContinentCode() string {
	return s.Continent
}

func (s GeoNameCountry) GetContinentName() string {
	return s.ContinentName
}

func (s GeoNameCountry) GetCountryCode() string {
	return s.Country.Iso2Code
}

func (s GeoNameCountry) GetCountryName() string {
	return s.GetName()
}

func (s GeoNameCountry) GetSubdivisionName() string {
	return ""
}

func (s GeoNameCountry) GetCityName() string {
	return ""
}

func (s GeoNameCountry) GetTimeZone() string {
	return ""
}

func (s GeoNameCountry) MarshalJSON() ([]byte, error) {
	return json.Marshal((*geoNameCountryJson)(s.Country))
}

func (s *GeoNameCountry) UnmarshalJSON(data []byte) error {
	var countryJson geoNameCountryJson
	if err := json.Unmarshal(data, &countryJson); err != nil {
		return err
	}
	s.Country = (*models.Country)(&countryJson)
	return nil
}
