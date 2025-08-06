package entity

import (
	"cmp"
	"encoding/json"
	"fmt"
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
	return cmp.Compare(time.Time(v).Unix(), time.Time(other).Unix())
}

func (v ModTimeVersion) MarshalJSON() ([]byte, error) {
	if time.Time(v).IsZero() {
		return json.Marshal(0)
	}
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
	DB    MMDBVersion     `json:"db,omitempty"`
	Patch *ModTimeVersion `json:"patch,omitempty"`
}

func (v PatchedMMDBVersion) Compare(other PatchedMMDBVersion) int {
	vPatch := cmp.Or(v.Patch, new(ModTimeVersion))
	oPatch := cmp.Or(other.Patch, new(ModTimeVersion))
	return cmp.Or(
		v.DB.Compare(other.DB),
		vPatch.Compare(*oPatch),
	)
}

func (v PatchedMMDBVersion) String() string {
	if v.Patch == nil {
		return fmt.Sprintf("%s-%d", v.DB.Version.String(), v.DB.BuildEpoch)
	}
	return fmt.Sprintf("%s-%d/p-%d", v.DB.Version.String(), v.DB.BuildEpoch, time.Time(*v.Patch).Unix())
}

type PatchedGeoNamesVersion struct {
	DB    GeoNamesVersion `json:"db,omitempty"`
	Patch *ModTimeVersion `json:"patch,omitempty"`
}

func (v PatchedGeoNamesVersion) Compare(other PatchedGeoNamesVersion) int {
	vPatch := cmp.Or(v.Patch, new(ModTimeVersion))
	oPatch := cmp.Or(other.Patch, new(ModTimeVersion))
	return cmp.Or(
		v.DB.Compare(other.DB),
		vPatch.Compare(*oPatch),
	)
}

func (v PatchedGeoNamesVersion) String() string {
	if v.Patch == nil {
		return fmt.Sprintf("%d", time.Time(v.DB).Unix())
	}
	return fmt.Sprintf("%d/p-%d", time.Time(v.DB).Unix(), time.Time(*v.Patch).Unix())
}

type Comparable[V any] = source.Comparable[V]

type Update[V source.Comparable[V]] = source.Update[V]

type Version[V any] interface {
	source.Comparable[V]
	fmt.Stringer
}

type DBUpdate[V Version[V]] struct {
	CurrentVersion   V       `json:"currentVersion"`
	AvailableVersion *V      `json:"availableVersion,omitempty"`
	UpdateError      *string `json:"updateError,omitempty"`
	InProgress       bool    `json:"inProgress,omitempty"`
}

func NewDBUpdate[V Version[V]](update Update[V], inProgress bool, lastUpdateError *string) DBUpdate[V] {
	res := DBUpdate[V]{
		CurrentVersion: update.CurrentVersion,
		InProgress:     inProgress,
	}

	if update.RemoteVersion.Compare(update.CurrentVersion) > 0 {
		res.AvailableVersion = &update.RemoteVersion
	}

	if !inProgress && lastUpdateError != nil {
		res.UpdateError = lastUpdateError
	}

	return res
}
