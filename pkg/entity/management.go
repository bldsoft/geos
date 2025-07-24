package entity

import "cmp"

type DBUpdate struct {
	DatabaseType    string `json:"databaseType"`
	Update          `json:",inline"`
	LastUpdateError string `json:"lastUpdateError,omitempty"`
}

func NewDBUpdate(dbType string, update Update, lastUpdateError *string) DBUpdate {
	res := DBUpdate{
		DatabaseType: dbType,
		Update:       update,
	}
	if !update.InProgress && update.AvailableVersion != "" && lastUpdateError != nil {
		res.LastUpdateError = *lastUpdateError
	}
	return res
}

type Update struct {
	CurrentVersion   string `json:"currentVersion"`
	AvailableVersion string `json:"availableVersion,omitempty"`
	InProgress       bool   `json:"inProgress,omitempty"`
}

func JoinUpdates(upd Update, updates ...Update) Update {
	for _, update := range updates {
		upd.CurrentVersion += "/" + update.CurrentVersion
		switch {
		case upd.AvailableVersion == "":
			upd.AvailableVersion = update.AvailableVersion
		case update.AvailableVersion != "":
			upd.AvailableVersion += "/" + update.AvailableVersion
		}
		upd.InProgress = cmp.Or(upd.InProgress, update.InProgress)
	}
	return upd
}
