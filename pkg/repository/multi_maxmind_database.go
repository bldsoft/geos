package repository

import (
	"bytes"
	"fmt"
	"io"
	"net"

	"github.com/bldsoft/geos/pkg/utils"
	"github.com/bldsoft/gost/log"
	"github.com/maxmind/mmdbwriter"
	"github.com/maxmind/mmdbwriter/inserter"
	"github.com/maxmind/mmdbwriter/mmdbtype"
	"github.com/oschwald/maxminddb-golang"
)

var ErrDBNotAvailable = fmt.Errorf("db %w", utils.ErrNotAvailable)

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

	// dbRaw, err := ioutil.ReadFile(path)
	var tree *mmdbwriter.Tree
	var err error
	log.Logger.WithFuncDuration(func() {

		tree, err = mmdbwriter.Load(path, mmdbwriter.Options{IncludeReservedNetworks: true})

	}).Info("loading " + path)
	if err != nil {
		handleErr(err)
		return nil
	}

	log.Logger.WithFuncDuration(func() {
		err = enrichDB(path, tree)
	}).Info("enriching " + path)
	if err != nil {
		handleErr(err)
		return nil
	}

	var buf bytes.Buffer
	if _, err := tree.WriteTo(&buf); err != nil {
		handleErr(err)
		return nil
	}
	dbRaw := buf.Bytes()

	// db, err := maxminddb.FromBytes(dbRaw)
	db, err := maxminddb.FromBytes(dbRaw)
	if err != nil {
		handleErr(err)
		return nil
	}

	return &database{
		path: path,
		db:   db,
		// dbRaw: dbRaw.Bytes(),
		dbRaw: dbRaw,
	}
}

func enrichDB(path string, tree *mmdbwriter.Tree) error {
	for _, priv := range []string{"10.10.0.0/8", "192.168.0.0/16", "172.16.0.0/12"} {
		_, private, _ := net.ParseCIDR(priv)
		log.Infof("Inserting %s", priv)
		data := mmdbtype.Map{
			"city":               mmdbtype.Map{"geoname_id": mmdbtype.Int32(9999999)},
			"continent":          mmdbtype.Map{"geoname_id": mmdbtype.Int32(9999999)},
			"country":            mmdbtype.Map{"geoname_id": mmdbtype.Int32(9999999)},
			"registered_country": mmdbtype.Map{"geoname_id": mmdbtype.Int32(9999999)},
			"location": mmdbtype.Map{
				"accuracy_radius": mmdbtype.Int32(1000),
				"latitude":        mmdbtype.Float64(32.751),
				"longitude":       mmdbtype.Float64(32.751),
				"time_zone":       mmdbtype.String("America/Chicago"),
			},
			"isp": mmdbtype.String("private network"),
		}
		if err := tree.InsertFunc(private, inserter.TopLevelMergeWith(data)); err != nil {
			return err
		}
	}
	return nil
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

func (db *database) RawData() (io.Reader, error) {
	if !db.Available() {
		return nil, ErrDBNotAvailable
	}
	return bytes.NewReader(db.dbRaw), nil
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
