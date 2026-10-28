package maxmind

import (
	"net"
	"net/netip"
)

func prefixToNetIPNet(prefix netip.Prefix) *net.IPNet {
	if !prefix.IsValid() {
		return nil
	}
	addr := prefix.Masked().Addr()
	return &net.IPNet{
		IP:   net.IP(addr.AsSlice()),
		Mask: net.CIDRMask(prefix.Bits(), addr.BitLen()),
	}
}
