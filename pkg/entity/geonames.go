package entity

import (
	"github.com/mkrou/geonames/models"
)

type GeoNameCountry struct {
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

type AdminSubdivision struct {
	Code      string `csv:"concatenated codes" valid:"required" json:"code"`
	Name      string `csv:"name" valid:"required" json:"name"`
	AsciiName string `csv:"asciiname" valid:"required" json:"asciiName"`
	GeonameId int    `csv:"geonameId" valid:"required" json:"geoNameID"`
}

type Geoname struct {
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
