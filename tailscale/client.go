package tailscale

import (
	"context"
	"fmt"
	"net/netip"
	"strings"

	"github.com/miekg/dns"
	"golang.org/x/crypto/ssh"
	"tailscale.com/client/local"
)

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
func (c *TSClient) QueryTSDNS(ctx context.Context, host string) (netip.Addr, error) {
	// Check if the host ends with the Tailnet suffix
	if !strings.HasSuffix(host, c.Tailnet) {
		host = host + "." + c.Tailnet
	}
	msg, _, err := c.Client.QueryDNS(ctx, host, "A")
	if err != nil {
		return netip.Addr{}, err
	}
	// Parse the DNS response
	dnsMsg := new(dns.Msg)
	if err := dnsMsg.Unpack(msg); err != nil {
		return netip.Addr{}, err
	}
	// Check for A records in the response
	for _, ans := range dnsMsg.Answer {
		if a, ok := ans.(*dns.A); ok {
			ip, _ := netip.AddrFromSlice(a.A)
			return ip, nil
		}
	}
	// If no A records found, return an error
	return netip.Addr{}, fmt.Errorf("no A record found for %s", host)
}

// GetSSHHostKeys retrieves the SSH host keys for the given IP address.
func (c *TSClient) GetSSHHostKeys(ctx context.Context, ip netip.Addr) ([]ssh.PublicKey, string, error) {
	// Use the WhoIs API to get the SSH host keys for the given IP address
	host, err := c.Client.WhoIs(ctx, ip.String())
	if err != nil {
		return nil, "", fmt.Errorf("failed to query WhoIs for %s: %w", ip, err)
	}
	if host == nil || !host.Node.Hostinfo.TailscaleSSHEnabled() {
		return nil, host.Node.Name, fmt.Errorf("Tailscale SSH is not enabled for %s", host.Node.Name)
	}

	// Parse the SSH host keys from the Hostinfo
	var keys []ssh.PublicKey
	for _, keyStr := range host.Node.Hostinfo.SSH_HostKeys().AsSlice() {
		key, _, _, _, err := ssh.ParseAuthorizedKey([]byte(keyStr))
		if err != nil {
			return nil, "", fmt.Errorf("failed to parse SSH host key for %s: %w", ip, err)
		}
		keys = append(keys, key)
	}
	return keys, host.Node.Name, nil
}

// Check IP is a Tailscale node
func (c *TSClient) IsTailscaleNode(ctx context.Context, ip netip.Addr) bool {
	return ipv4_prefix.Contains(ip) || ipv6_prefix.Contains(ip)
}

// GetHost returns the Tailscale host information for the given IP address.
func (c *TSClient) GetHost(ctx context.Context, host string) (*TailscaleHost, error) {
	ip, err := netip.ParseAddr(host)
	if err != nil {
		// The host is not an IP, assume it's a hostname
		if !strings.HasSuffix(host, c.Tailnet) {
			host = host + "." + c.Tailnet
		}
		ip, err = c.QueryTSDNS(ctx, host)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve %s: %w", host, err)
		}
	}
	if !c.IsTailscaleNode(ctx, ip) {
		return nil, fmt.Errorf("%s is not a Tailscale node", host)
	}

	keys, host, err := c.GetSSHHostKeys(ctx, ip)
	if err != nil {
		return nil, fmt.Errorf("failed to get SSH host keys for %s: %w", host, err)
	}

	keysMap := make(map[string]ssh.PublicKey)
	for _, key := range keys {
		keyType := key.Type()
		keysMap[keyType] = key
	}

	return &TailscaleHost{
		Name: host,
		IP:   ip,
		Keys: keysMap,
	}, nil
}
