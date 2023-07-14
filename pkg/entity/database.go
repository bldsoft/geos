package entity

import (
	"strings"

	"github.com/oschwald/maxminddb-golang"
)

type MetaData = maxminddb.Metadata

type Database struct {
	Data []byte
	MetaData
	Ext string
}

func (db *Database) FileName() string {
	return db.DatabaseType + "." + strings.TrimLeft(db.Ext, ".")
}
