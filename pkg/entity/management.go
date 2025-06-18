package entity

type Subject string

const (
	SubjectISPDb           Subject = "ISP"
	SubjectCitiesDb        Subject = "City"
	SubjectGeonames        Subject = "Geonames"
	SubjectISPDbPatches    Subject = "ISP patches"
	SubjectCitiesDbPatches Subject = "City patches"
	SubjectGeonamesPatches Subject = "Geonames patches"
)

type Updates map[Subject]*UpdateStatus

type UpdateStatus struct {
	Error     string `json:"error,omitempty"` //source interaction errors
	Available bool   `json:"available"`
}
