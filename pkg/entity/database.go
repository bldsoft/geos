package entity

import (
	"io"
	"strings"

	"github.com/oschwald/maxminddb-golang"
)

type MetaData = maxminddb.Metadata

type Database struct {
	Data io.Reader
	MetaData
	Ext string
}

func (db *Database) FileName() string {
	return db.DatabaseType + "." + strings.TrimLeft(db.Ext, ".")
}
