package entity

import (
	"cmp"
	"encoding/json"
	"time"

	"github.com/bldsoft/geos/pkg/storage/source"

	"github.com/hashicorp/go-version"
)

type MMDBVersion source.MMDBVersion

func (v MMDBVersion) Compare(other MMDBVersion) int {
	return source.MMDBVersion(v).Compare(source.MMDBVersion(other))
}

func (v MMDBVersion) MarshalJSON() ([]byte, error) {
	type tmp struct {
		Version    string `json:"version,omitempty"`
		BuildEpoch uint   `json:"buildEpoch,omitempty"`
	}
	if v.Version == nil {
		return json.Marshal(tmp{})
	}
	return json.Marshal(tmp{
		Version:    v.Version.String(),
		BuildEpoch: v.BuildEpoch,
	})
}

func (v *MMDBVersion) UnmarshalJSON(data []byte) error {
	type tmp struct {
		Version    string `json:"version"`
		BuildEpoch uint   `json:"buildEpoch"`
	}
	var tmpVersion tmp
	if err := json.Unmarshal(data, &tmpVersion); err != nil {
		return err
	}
	var err error
	v.Version, err = version.NewVersion(tmpVersion.Version)
	if err != nil {
		return err
	}
	v.BuildEpoch = tmpVersion.BuildEpoch
	return nil
}

type ModTimeVersion time.Time

func (v ModTimeVersion) Compare(other ModTimeVersion) int {
	return time.Time(v).Compare(time.Time(other))
}

func (v ModTimeVersion) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(v).Unix())
}

func (v *ModTimeVersion) UnmarshalJSON(data []byte) error {
	var tmp int64
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	*v = ModTimeVersion(time.Unix(tmp, 0))
	return nil
}

type PatchVersion = ModTimeVersion
type GeoNamesVersion = ModTimeVersion

type PatchedMMDBVersion struct {
	DB    MMDBVersion    `json:"db,omitempty"`
	Patch ModTimeVersion `json:"patch,omitempty"`
}

func (v PatchedMMDBVersion) Compare(other PatchedMMDBVersion) int {
	return cmp.Or(
		v.DB.Compare(other.DB),
		v.Patch.Compare(other.Patch),
	)
}

type PatchedGeoNamesVersion struct {
	DB           GeoNamesVersion
	PatchVersion ModTimeVersion
}

func (v PatchedGeoNamesVersion) Compare(other PatchedGeoNamesVersion) int {
	return cmp.Or(
		v.DB.Compare(other.DB),
		v.PatchVersion.Compare(other.PatchVersion),
	)
}

type Update[V source.Comparable[V]] = source.Update[V]

type DBUpdate[V source.Comparable[V]] struct {
	CurrentVersion   V      `json:"currentVersion"`
	AvailableVersion *V     `json:"availableVersion,omitempty"`
	UpdateError      string `json:"updateError,omitempty"`
	InProgress       bool   `json:"inProgress,omitempty"`
}

func NewDBUpdate[V source.Comparable[V]](update Update[V], inProgress bool, lastUpdateError *string) DBUpdate[V] {
	res := DBUpdate[V]{
		CurrentVersion: update.CurrentVersion,
		InProgress:     inProgress,
	}

	if update.RemoteVersion.Compare(update.CurrentVersion) > 0 {
		res.AvailableVersion = &update.RemoteVersion
	}

	if !inProgress && lastUpdateError != nil {
		res.UpdateError = *lastUpdateError
	}

	return res
}
