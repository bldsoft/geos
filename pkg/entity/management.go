package entity

import "encoding/json"

type DBUpdate struct {
	DatabaseType     string `json:"databaseType"`
	CurrentVersion   string `json:"currentVersion"`
	AvailableVersion string `json:"availableVersion,omitempty"`
	LastUpdateError  string `json:"lastUpdateError,omitempty"`
	InProgress       bool   `json:"inProgress,omitempty"`
}

func NewDBUpdate(dbType string, update Update, inProgress bool, lastUpdateError *string) DBUpdate {
	res := DBUpdate{
		DatabaseType:     dbType,
		CurrentVersion:   update.CurrentVersion,
		AvailableVersion: update.AvailableVersion,
		InProgress:       inProgress,
	}
	if !inProgress && lastUpdateError != nil {
		res.LastUpdateError = *lastUpdateError
	}
	return res
}

func (u DBUpdate) MarshalJSON() ([]byte, error) {
	if u.AvailableVersion == u.CurrentVersion {
		u.AvailableVersion = ""
	}
	type tmp DBUpdate
	return json.Marshal(tmp(u))
}

type Update struct {
	CurrentVersion   string
	AvailableVersion string
}

func JoinUpdates(upd Update, updates ...Update) Update {
	for _, update := range updates {
		upd.CurrentVersion = joinVersion(upd.CurrentVersion, update.CurrentVersion)
		upd.AvailableVersion = joinVersion(upd.AvailableVersion, update.AvailableVersion)
	}
	return upd
}

func joinVersion(version1, version2 string) string {
	if version1 == "" {
		return version2
	}
	if version2 == "" {
		return version1
	}
	return version1 + "/" + version2
}
