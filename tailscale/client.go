package tailscale

import (
	"context"
	"net"

	"tailscale.com/client/local"
	"tailscale.com/ipn/ipnstate"
	"tailscale.com/types/dnstype"
)

type Client interface {
	Status(ctx context.Context) (*ipnstate.Status, error)
	QueryDNS(ctx context.Context, host string, qtype string) ([]byte, []*dnstype.Resolver, error)
}

type TSClient struct {
	Client
	tailnet string
}

func NewTSClient() (*TSClient, error) {
	ctx := context.TODO()
	client := &TSClient{
		Client: &local.Client{},
	}
	status, err := client.Client.Status(ctx)
	if err != nil {
		return nil, err
	}
	client.tailnet = status.CurrentTailnet.MagicDNSSuffix
	return client, nil
}

// QueryDNS queries the Tailscale DNS for the given host.
// It returns the IP address if found, or an error if not found.
func (c *TSClient) QueryDNS(ctx context.Context, host string) (*net.IP, error) {
	return nil, nil
}
