package entity

import "errors"

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
	Error     string `json:"error,omitempty"` //external errors during update/update check
	Available bool   `json:"available"`
}

func (u *Updates) Error() error {
	var multiErr error
	for _, status := range *u {
		if status.Error != "" {
			multiErr = errors.Join(multiErr, errors.New(status.Error))
		}
	}
	return multiErr
}
