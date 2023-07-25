package repository

import (
	"errors"
	"io/ioutil"
	"net"

	"github.com/bldsoft/gost/log"
	"github.com/oschwald/maxminddb-golang"
)

var ErrDBNotAvailable = errors.New("db not available")

type database struct {
	path  string
	db    *maxminddb.Reader
	dbRaw []byte
}

func openDB(path string, dbType MaxmindDBType, required bool) *database {
	handleErr := func(err error) {
		if err != nil {
			if required {
				log.Fatalf("Failed to read %s db: %s", dbType, err)
			}
			log.Warnf("Failed to read %s db: %s", dbType, err)
		}
	}
	dbRaw, err := ioutil.ReadFile(path)
	if err != nil {
		handleErr(err)
		return nil
	}

	db, err := maxminddb.FromBytes(dbRaw)
	if err != nil {
		handleErr(err)
		return nil
	}

	return &database{
		path:  path,
		db:    db,
		dbRaw: dbRaw,
	}
}

func (db *database) Available() bool {
	return db != nil
}

func (db *database) Path() (string, error) {
	if !db.Available() {
		return "", ErrDBNotAvailable
	}
	return db.path, nil
}

func (db *database) RawData() ([]byte, error) {
	if !db.Available() {
		return nil, ErrDBNotAvailable
	}
	return db.dbRaw, nil
}

func (db *database) MetaData() (*maxminddb.Metadata, error) {
	if !db.Available() {
		return nil, ErrDBNotAvailable
	}
	return &db.db.Metadata, nil
}

func (db *database) Close() error {
	if !db.Available() {
		return ErrDBNotAvailable
	}
	return db.db.Close()
}

func (db *database) Decode(offset uintptr, result interface{}) error {
	if !db.Available() {
		return ErrDBNotAvailable
	}
	return db.db.Decode(offset, result)
}

func (db *database) Lookup(ip net.IP, result interface{}) error {
	if !db.Available() {
		return ErrDBNotAvailable
	}
	return db.db.Lookup(ip, result)
}

func (db *database) LookupNetwork(ip net.IP, result interface{}) (network *net.IPNet, ok bool, err error) {
	if !db.Available() {
		return nil, false, ErrDBNotAvailable
	}
	return db.db.LookupNetwork(ip, result)
}

func (db *database) LookupOffset(ip net.IP) (uintptr, error) {
	if !db.Available() {
		return 0, ErrDBNotAvailable
	}
	return db.db.LookupOffset(ip)
}

func (db *database) Networks(options ...maxminddb.NetworksOption) (*maxminddb.Networks, error) {
	if !db.Available() {
		return nil, ErrDBNotAvailable
	}
	return db.db.Networks(), nil
}

func (db *database) NetworksWithin(network *net.IPNet, options ...maxminddb.NetworksOption) (*maxminddb.Networks, error) {
	if !db.Available() {
		return nil, ErrDBNotAvailable
	}
	return db.db.NetworksWithin(network, options...), nil
}

func (db *database) Verify() error {
	if !db.Available() {
		return ErrDBNotAvailable
	}
	return db.db.Verify()
}
