package entity

import (
	"encoding/json"
	"strings"

	"github.com/mkrou/geonames/models"
)

type GeoNameAdminSubdivision struct {
	*models.AdminDivision
	ContinentCode string `csv:"continent code"`
	ContinentName string `csv:"continent name"`
	CountryName   string `csv:"country name"`
}

type adminSubdivisionJson struct {
	*modelSubdivisionJson
	ContinentCode string `csv:"continent code" json:"continentCode"`
	ContinentName string `csv:"continent name" json:"continentName"`
	CountryName   string `csv:"country name" json:"countryName"`
}

// same as models.AdminDivision, but with json tags
type modelSubdivisionJson struct {
	Code      string `csv:"concatenated codes" valid:"required" json:"code"`
	Name      string `csv:"name" valid:"required" json:"name"`
	AsciiName string `csv:"asciiname" valid:"required" json:"asciiName"`
	GeonameId int    `csv:"geonameId" valid:"required" json:"geoNameID"`
}

func (s GeoNameAdminSubdivision) GetGeoNameID() int {
	return s.AdminDivision.GeonameId
}

func (s GeoNameAdminSubdivision) GetName() string {
	return s.AdminDivision.Name
}

func (s GeoNameAdminSubdivision) AdminCode() string {
	splitted := strings.SplitN(s.AdminDivision.Code, ".", 2)
	if len(splitted) < 2 {
		return splitted[0]
	}
	return splitted[1]
}

func (s GeoNameAdminSubdivision) GetContinentCode() string {
	return s.ContinentCode
}

func (s GeoNameAdminSubdivision) GetContinentName() string {
	return s.ContinentName
}

func (s GeoNameAdminSubdivision) GetCountryCode() string {
	return strings.SplitN(s.AdminDivision.Code, ".", 2)[0]
}

func (s GeoNameAdminSubdivision) GetCountryName() string {
	return s.CountryName
}

func (s GeoNameAdminSubdivision) GetSubdivisionName() string {
	return s.GetName()
}

func (s GeoNameAdminSubdivision) GetCityName() string {
	return ""
}

func (s GeoNameAdminSubdivision) GetTimeZone() string {
	return ""
}

func (s GeoNameAdminSubdivision) MarshalJSON() ([]byte, error) {
	return json.Marshal(&adminSubdivisionJson{
		modelSubdivisionJson: (*modelSubdivisionJson)(s.AdminDivision),
		ContinentCode:        s.ContinentCode,
		ContinentName:        s.ContinentName,
		CountryName:          s.CountryName,
	})
}

func (s *GeoNameAdminSubdivision) UnmarshalJSON(data []byte) error {
	var subdivisionJson adminSubdivisionJson
	if err := json.Unmarshal(data, &subdivisionJson); err != nil {
		return err
	}
	s.AdminDivision = (*models.AdminDivision)(subdivisionJson.modelSubdivisionJson)
	s.ContinentCode = subdivisionJson.ContinentCode
	s.ContinentName = subdivisionJson.ContinentName
	s.CountryName = subdivisionJson.CountryName
	return nil
}
