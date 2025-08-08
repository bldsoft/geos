package maxmind

import (
	"context"
	"io"
	"net"

	"github.com/oschwald/maxminddb-golang"
)

type Database interface {
	Lookup(ctx context.Context, ip net.IP, result interface{}) error
	// LookupNetwork(ip net.IP, result interface{}) (network *net.IPNet, ok bool, err error)
	// LookupOffset(ip net.IP) (uintptr, error)
	Networks(ctx context.Context, options ...maxminddb.NetworksOption) (*maxminddb.Networks, error)
	// NetworksWithin(network *net.IPNet, options ...maxminddb.NetworksOption) (*maxminddb.Networks, error)
	// Verify() error
	// Close() error

	RawData(ctx context.Context) (io.Reader, error) // mmdb
	MetaData(ctx context.Context) (*maxminddb.Metadata, error)
}

type CSVDumper interface {
	Database
	WriteCSVTo(ctx context.Context, w io.Writer) error
	CSV(ctx context.Context, gzipCompress bool) (io.Reader, error)
}

type CSVEntity interface {
	MarshalCSV() (names, row []string, err error)
}
