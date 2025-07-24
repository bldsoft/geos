package maxmind

import (
	"context"
	"io"
	"net"

	"github.com/bldsoft/geos/pkg/storage/source"
	"github.com/oschwald/maxminddb-golang"
)

type Database interface {
	Lookup(ip net.IP, result interface{}) error
	// LookupNetwork(ip net.IP, result interface{}) (network *net.IPNet, ok bool, err error)
	// LookupOffset(ip net.IP) (uintptr, error)
	Networks(options ...maxminddb.NetworksOption) (*maxminddb.Networks, error)
	// NetworksWithin(network *net.IPNet, options ...maxminddb.NetworksOption) (*maxminddb.Networks, error)
	// Verify() error
	// Close() error

	RawData() (io.Reader, error) // mmdb
	MetaData() (*maxminddb.Metadata, error)

	source.RecoverableUpdater
}

type CSVDumper interface {
	Database
	WriteCSVTo(ctx context.Context, w io.Writer) error
	CSV(ctx context.Context, gzipCompress bool) (io.Reader, error)
}

type CSVEntity interface {
	MarshalCSV() (names, row []string, err error)
}
