package entity

import (
	"encoding/json"

	"github.com/bldsoft/geos/pkg/storage/source"
)

type Version interface {
	Compare(Version) int
}

type MMDBVersion = source.MMDBVersion
type ModTimeVersion = source.ModTimeVersion
type GeoNamesVersion = source.ModTimeVersion
type PatchVersion = source.ModTimeVersion

type PatchedMMDBVersion struct {
	MMDBVersion
	PatchVersion PatchVersion
}

type PatchedGeoNamesVersion struct {
	GeoNamesVersion
	PatchVersion PatchVersion
}

func (v PatchedGeoNamesVersion) IsHigher(other PatchedGeoNamesVersion) bool {
	return v.GeoNamesVersion.IsHigher(other.GeoNamesVersion) || v.PatchVersion.IsHigher(other.PatchVersion)
}

type DBUpdate[V Version[V]] struct {
	CurrentVersion   V      `json:"currentVersion"`
	AvailableVersion V      `json:"availableVersion,omitempty"`
	UpdateError      string `json:"updateError,omitempty"`
	InProgress       bool   `json:"inProgress,omitempty"`
}

func NewDBUpdate[V Version[V]](update Update[V], inProgress bool, lastUpdateError *string) DBUpdate[V] {
	res := DBUpdate[V]{
		CurrentVersion:   update.CurrentVersion,
		AvailableVersion: update.RemoteVersion,
		InProgress:       inProgress,
	}

	if !inProgress && lastUpdateError != nil {
		res.UpdateError = *lastUpdateError
	}

	return res
}

func (u DBUpdate[V]) MarshalJSON() ([]byte, error) {
	if !u.AvailableVersion.IsHigher(u.CurrentVersion) {
		var zero V
		u.AvailableVersion = zero
	}
	type tmp DBUpdate[V]
	return json.Marshal(tmp(u))
}
