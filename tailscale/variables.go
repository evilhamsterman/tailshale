package tailscale

import (
	"context"
	"net/netip"

	"tailscale.com/client/tailscale/apitype"
	"tailscale.com/ipn/ipnstate"
	"tailscale.com/types/dnstype"
)

// Supported key types for Tailscale SSH
const (
	RSA     = "ssh-rsa"
	ECDSA   = "ecdsa-sha2-nistp256"
	ED25519 = "ssh-ed25519"
)

var (
	// Tailscale IP address prefix for IPv4
	ipv4_prefix = netip.MustParsePrefix("100.64.0.0/10")
	// Tailscale IP address prefix for IPv6
	ipv6_prefix = netip.MustParsePrefix("fd7a:115c:a1e0::/48")
)

// Client is an interface that defines the methods we need for interacting with
// Tailscale's local client API.
type Client interface {
	Status(ctx context.Context) (*ipnstate.Status, error)
	QueryDNS(ctx context.Context, host string, qtype string) ([]byte, []*dnstype.Resolver, error)
	WhoIs(ctx context.Context, ip string) (*apitype.WhoIsResponse, error)
}

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
