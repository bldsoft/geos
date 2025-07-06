package state

import (
	"encoding/json"
	"go/version"
)

type GeosState struct {
	CityVersion              string
	CityPatchesTimestamp     int64
	ISPVersion               string
	ISPPatchesTimestamp      int64
	GeonamesTimestamps       int64
	GeonamesPatchesTimestamp int64
}

func (s *GeosState) Add(other *GeosState) {
	if s.CityVersion == "" && other.CityVersion != "" {
		s.CityVersion = other.CityVersion
	}
	if s.CityPatchesTimestamp == 0 && other.CityPatchesTimestamp != 0 {
		s.CityPatchesTimestamp = other.CityPatchesTimestamp
	}
	if s.ISPVersion == "" && other.ISPVersion != "" {
		s.ISPVersion = other.ISPVersion
	}
	if s.ISPPatchesTimestamp == 0 && other.ISPPatchesTimestamp != 0 {
		s.ISPPatchesTimestamp = other.ISPPatchesTimestamp
	}
	s.GeonamesTimestamps += other.GeonamesTimestamps
	if s.GeonamesPatchesTimestamp == 0 && other.GeonamesPatchesTimestamp != 0 {
		s.GeonamesPatchesTimestamp = other.GeonamesPatchesTimestamp
	}
}

func (s *GeosState) IsHigher(other *GeosState) bool {
	if s == nil {
		return false
	}

	if other == nil {
		return true
	}

	if version.Compare(s.CityVersion, other.CityVersion) > 0 {
		return true
	}
	if version.Compare(s.ISPVersion, other.ISPVersion) > 0 {
		return true
	}

	if s.CityPatchesTimestamp > other.CityPatchesTimestamp {
		return true
	}
	if s.ISPPatchesTimestamp > other.ISPPatchesTimestamp {
		return true
	}

	if s.GeonamesTimestamps > other.GeonamesTimestamps {
		return true
	}
	if s.GeonamesPatchesTimestamp > other.GeonamesPatchesTimestamp {
		return true
	}

	return false
}

func (s *GeosState) Parse(versionStr string) (*GeosState, error) {
	var state GeosState
	err := json.Unmarshal([]byte(versionStr), &state)
	if err != nil {
		return &GeosState{}, nil
	}
	return &state, nil
}

func Parse(versionStr string) (*GeosState, error) {
	var state GeosState
	err := json.Unmarshal([]byte(versionStr), &state)
	if err != nil {
		return &GeosState{}, nil
	}
	return &state, nil
}
