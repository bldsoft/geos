package entity

import (
	"io"
	"strings"

	"github.com/oschwald/maxminddb-golang/v2"
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
