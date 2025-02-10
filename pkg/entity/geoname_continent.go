package entity

import "encoding/json"

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

func (c GeoNameContinent) GetGeoNameID() int {
	return c.geonameID
}

func (c GeoNameContinent) GetName() string {
	return c.name
}

func (c GeoNameContinent) Code() string {
	return c.code
}

func (c GeoNameContinent) GetContinentCode() string {
	return c.Code()
}

func (c GeoNameContinent) GetContinentName() string {
	return c.GetName()
}

func (c GeoNameContinent) GetCountryCode() string {
	return ""
}

func (c GeoNameContinent) GetCountryName() string {
	return ""
}

func (c GeoNameContinent) GetSubdivisionName() string {
	return ""
}

func (c GeoNameContinent) GetCityName() string {
	return ""
}

func (c GeoNameContinent) GetTimeZone() string {
	return ""
}

func (s GeoNameContinent) MarshalJSON() ([]byte, error) {
	return json.Marshal(geoNameContinentJson{
		GeonameID: s.GetGeoNameID(),
		Code:      s.Code(),
		Name:      s.GetName(),
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
