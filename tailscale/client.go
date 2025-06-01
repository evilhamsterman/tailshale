package tailscale

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/miekg/dns"
	"tailscale.com/client/local"
	"tailscale.com/ipn/ipnstate"
	"tailscale.com/types/dnstype"
)

// Client is an interface that defines the methods we need for interacting with
// Tailscale's local client API.
type Client interface {
	Status(ctx context.Context) (*ipnstate.Status, error)
	QueryDNS(ctx context.Context, host string, qtype string) ([]byte, []*dnstype.Resolver, error)
}

var _ Client = (*local.Client)(nil) // Ensure the tailscale local.Client implements the Client interface

type TSClient struct {
	Client  Client
	Tailnet string
}

func NewTSClient(c Client) (*TSClient, error) {
	ctx := context.TODO()
	client := &TSClient{
		Client: c,
	}
	status, err := client.Client.Status(ctx)
	if err != nil {
		return nil, err
	}
	client.Tailnet = status.CurrentTailnet.MagicDNSSuffix
	return client, nil
}

// QueryTSDNS queries the Tailscale DNS for the given host.
// It returns the IP address if found, or an error if not found.
func (c *TSClient) QueryTSDNS(ctx context.Context, host string) (*net.IP, error) {
	// Check if the host ends with the Tailnet suffix
	if !strings.HasSuffix(host, c.Tailnet) {
		host = host + "." + c.Tailnet
	}
	msg, _, err := c.Client.QueryDNS(ctx, host, "A")
	if err != nil {
		return nil, err
	}
	// Parse the DNS response
	dnsMsg := new(dns.Msg)
	if err := dnsMsg.Unpack(msg); err != nil {
		return nil, err
	}
	// Check for A records in the response
	for _, ans := range dnsMsg.Answer {
		if a, ok := ans.(*dns.A); ok {
			ip := net.ParseIP(a.A.String())
			if ip != nil {
				return &ip, nil
			}
		}
	}
	// If no A records found, return an error
	return nil, fmt.Errorf("no A record found for %s", host)
}
