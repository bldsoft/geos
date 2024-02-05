package entity

import (
	"encoding/json"
	"slices"
	"strings"

	"github.com/mkrou/geonames/models"
)

type GeoNameEntity interface {
	GeoNameID() int
	Name() string
	CountryCode() string
}

type GeoNameCountry struct {
	*models.Country
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

func (s GeoNameCountry) GeoNameID() int {
	return s.Country.GeonameID
}

func (s GeoNameCountry) Name() string {
	return s.Country.Name
}

func (s GeoNameCountry) CountryCode() string {
	return s.Country.Iso2Code
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

type GeoNameAdminSubdivision struct {
	*models.AdminDivision
}

// same as models.AdminDivision, but with json tags
type adminSubdivisionJson struct {
	Code      string `csv:"concatenated codes" valid:"required" json:"code"`
	Name      string `csv:"name" valid:"required" json:"name"`
	AsciiName string `csv:"asciiname" valid:"required" json:"asciiName"`
	GeonameId int    `csv:"geonameId" valid:"required" json:"geoNameID"`
}

func (s GeoNameAdminSubdivision) GeoNameID() int {
	return s.AdminDivision.GeonameId
}

func (s GeoNameAdminSubdivision) Name() string {
	return s.AdminDivision.Name
}

func (s GeoNameAdminSubdivision) AdminCode() string {
	if len(s.AdminDivision.Code) < 3 {
		return ""
	}
	return s.AdminDivision.Code[3:]
}

func (s GeoNameAdminSubdivision) CountryCode() string {
	if len(s.AdminDivision.Code) < 2 {
		return ""
	}
	return s.AdminDivision.Code[:2]
}

func (s GeoNameAdminSubdivision) MarshalJSON() ([]byte, error) {
	return json.Marshal((*adminSubdivisionJson)(s.AdminDivision))
}

func (s *GeoNameAdminSubdivision) UnmarshalJSON(data []byte) error {
	var subdivisionJson adminSubdivisionJson
	if err := json.Unmarshal(data, &subdivisionJson); err != nil {
		return err
	}
	s.AdminDivision = (*models.AdminDivision)(&subdivisionJson)
	return nil
}

type GeoName struct {
	*models.Geoname
}

// same as models.Geoname, but with json tags
type geoNameJson struct {
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

func (g GeoName) GeoNameID() int {
	return g.Geoname.Id
}

func (g GeoName) Name() string {
	return g.Geoname.Name
}

func (g GeoName) CountryCode() string {
	return g.Geoname.CountryCode
}

func (s GeoName) MarshalJSON() ([]byte, error) {
	return json.Marshal((*geoNameJson)(s.Geoname))
}

func (s *GeoName) UnmarshalJSON(data []byte) error {
	var geoNameJson geoNameJson
	if err := json.Unmarshal(data, &geoNameJson); err != nil {
		return err
	}
	s.Geoname = (*models.Geoname)(&geoNameJson)
	return nil
}

type GeoNameFilter struct {
	GeoNameIDs   []uint32 `schema:"geoname-ids"`
	CountryCodes []string `schema:"country-codes"`
	NamePrefix   string   `schema:"name-prefix"`
	Limit        uint32   `schema:"limit"`
}

func (f *GeoNameFilter) Match(e GeoNameEntity) bool {
	if len(f.GeoNameIDs) > 0 && !slices.Contains(f.GeoNameIDs, uint32(e.GeoNameID())) {
		return false
	}

	if len(f.CountryCodes) > 0 && !slices.Contains(f.CountryCodes, e.CountryCode()) {
		return false
	}

	if len(f.NamePrefix) > 0 && !strings.HasPrefix(e.Name(), f.NamePrefix) {
		return false
	}

	return true
}

type GeoNameContinent struct {
	geonameID int
	code      string
	name      string
}

type geoNameContinentJson struct {
	GeonameID int    `valid:"required" json:"geoNameID"`
	Code      string `valid:"required" json:"code"`
	Name      string `valid:"required" json:"name"`
}

func NewGeoNameContinent(geonameID int, code, name string) *GeoNameContinent {
	return &GeoNameContinent{
		geonameID: geonameID,
		code:      code,
		name:      name,
	}
}

func (c GeoNameContinent) GeoNameID() int {
	return c.geonameID
}

func (c GeoNameContinent) Name() string {
	return c.name
}

func (c GeoNameContinent) Code() string {
	return c.code
}

func (c GeoNameContinent) CountryCode() string {
	return ""
}

func (s GeoNameContinent) MarshalJSON() ([]byte, error) {
	return json.Marshal(geoNameContinentJson{
		GeonameID: s.GeoNameID(),
		Code:      s.Code(),
		Name:      s.Name(),
	})
}

func (s *GeoNameContinent) UnmarshalJSON(data []byte) error {
	var geoNameContinentJson geoNameContinentJson
	if err := json.Unmarshal(data, &geoNameContinentJson); err != nil {
		return err
	}

	s.geonameID = geoNameContinentJson.GeonameID
	s.code = geoNameContinentJson.Code
	s.name = geoNameContinentJson.Name
	return nil
}

var _ GeoNameEntity = &GeoNameCountry{}
var _ GeoNameEntity = &GeoNameAdminSubdivision{}
var _ GeoNameEntity = &GeoName{}
var _ GeoNameEntity = &GeoNameContinent{}
