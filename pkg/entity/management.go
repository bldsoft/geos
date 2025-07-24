package entity

import (
	"encoding/json"
)

type DBUpdate struct {
	DatabaseType     string `json:"databaseType"`
	CurrentVersion   string `json:"currentVersion"`
	AvailableVersion string `json:"availableVersion,omitempty"`
	UpdateError      string `json:"updateError,omitempty"`
	InProgress       bool   `json:"inProgress,omitempty"`
}

func NewDBUpdate(dbType string, update Update, inProgress bool, lastUpdateError *string) DBUpdate {
	res := DBUpdate{
		DatabaseType:     dbType,
		CurrentVersion:   update.CurrentVersion,
		AvailableVersion: update.RemoteVersion,
		InProgress:       inProgress,
	}

	if !inProgress && lastUpdateError != nil {
		res.UpdateError = *lastUpdateError
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
	CurrentVersion string
	RemoteVersion  string
}

func JoinUpdates(upd Update, updates ...Update) Update {
	for _, update := range updates {
		upd.CurrentVersion = joinString(upd.CurrentVersion, update.CurrentVersion, "/")
		upd.RemoteVersion = joinString(upd.RemoteVersion, update.RemoteVersion, "/")
	}
	return upd
}

func joinString(s1, s2 string, sep string) string {
	if s1 == "" {
		return s2
	}
	if s2 == "" {
		return s1
	}
	return s1 + sep + s2
}
