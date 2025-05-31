package tailscale

import (
	"net/netip"
)

var (
	// Tailscale IP address prefix for IPv4
	ipv4_prefix = netip.MustParsePrefix("100.64.0.0/10")
	// Tailscale IP address prefix for IPv6
	ipv6_prefix = netip.MustParsePrefix("fd7a:115c:a1e0::/48")
)

type InvalideTailscaleNameError struct {
	Host string
}

func (e *InvalideTailscaleNameError) Error() string {
	return "invalid Tailscale name: " + e.Host
}

type InvalidTailscaleIPError struct {
	IP netip.Addr
}

func (e *InvalidTailscaleIPError) Error() string {
	return "invalid Tailscale IP address: " + e.IP.String()
}
