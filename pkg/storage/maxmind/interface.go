package maxmind

import (
	"context"
	"io"
	"iter"
	"net/netip"

	"github.com/oschwald/maxminddb-golang/v2"
)

type Database interface {
	Lookup(ctx context.Context, ip netip.Addr, result any) error
	// LookupNetwork(ip netip.Addr, result interface{}) (network netip.Prefix, ok bool, err error)
	// LookupOffset(ip netip.Addr) (uintptr, error)
	Networks(ctx context.Context, options ...maxminddb.NetworksOption) (iter.Seq[maxminddb.Result], error)
	// NetworksWithin(network netip.Prefix, options ...maxminddb.NetworksOption) (*maxminddb.Networks, error)
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
