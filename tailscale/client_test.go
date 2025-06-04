package tailscale

import (
	"context"
	"net/netip"
	"testing"

	in "github.com/evilhamsterman/tailshale/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"tailscale.com/client/tailscale/apitype"
	"tailscale.com/ipn/ipnstate"
	"tailscale.com/types/dnstype"
)

var _ Client = (*in.MockClient)(nil) // Ensure MockClient implements the Client interface

func TestNewTSClient(t *testing.T) {
	m := new(in.MockClient)
	m.On("Status", mock.Anything).Return(&ipnstate.Status{
		CurrentTailnet: &ipnstate.TailnetStatus{
			MagicDNSSuffix: in.TEST_TAILNET,
		},
	}, nil)

	client, err := NewTSClient(m)
	m.AssertExpectations(t)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, in.TEST_TAILNET, client.Tailnet)
}

func TestQueryDNS(t *testing.T) {
	msg := in.GetTestDNSMessage()

	m := new(in.MockClient)
	m.On("QueryDNS", context.TODO(), "test.example.ts.net", "A").Return(
		msg, []*dnstype.Resolver{}, nil)

	c := &TSClient{
		Client:  m,
		Tailnet: in.TEST_TAILNET,
	}
	ip, err := c.QueryTSDNS(context.TODO(), "test.example.ts.net")
	m.AssertExpectations(t)
	assert.NoError(t, err)
	assert.NotNil(t, ip)
	assert.Equal(t, netip.MustParseAddr("100.100.100.100"), ip)
}

func TestGetSSHHostKeys(t *testing.T) {
	m := new(in.MockClient)
	m.On("WhoIs", context.TODO(), in.TEST_IP.String()).Return(
		&apitype.WhoIsResponse{
			Node: in.GetTestNode([]string{in.TEST_HOST_KEY})},
		nil)

	c := &TSClient{
		Client:  m,
		Tailnet: in.TEST_TAILNET,
	}
	host, err := c.GetSSHHostKeys(context.TODO(), in.TEST_IP)
	m.AssertExpectations(t)
	assert.NoError(t, err)
	assert.Equal(t, "test."+in.TEST_TAILNET, host.Name)
	assert.Equal(t, in.TEST_IP, host.IP)
	require.Len(t, host.Keys, 1)
	assert.Equal(t, in.TEST_HOST_KEY_OBJECT, host.Keys[ED25519])
}

func TestGetSSHHostKeys_NoSSH(t *testing.T) {
	m := new(in.MockClient)
	m.On("WhoIs", context.TODO(), in.TEST_IP.String()).Return(
		&apitype.WhoIsResponse{
			Node: in.GetTestNode(nil)},
		nil)

	c := &TSClient{
		Client:  m,
		Tailnet: in.TEST_TAILNET,
	}
	host, err := c.GetSSHHostKeys(context.TODO(), in.TEST_IP)
	m.AssertExpectations(t)
	assert.Error(t, err)
	assert.Equal(t, "test."+in.TEST_TAILNET, host.Name)
	assert.Nil(t, host.Keys)
}

func TestIsTailcaleNode(t *testing.T) {
	c := &TSClient{
		Tailnet: in.TEST_TAILNET,
	}
	tests := []struct {
		ip       netip.Addr
		expected bool
	}{
		{
			ip:       in.TEST_IP,
			expected: true,
		},
		{
			ip:       netip.MustParseAddr("192.168.0.1"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.ip.String(), func(t *testing.T) {
			isNode := c.IsTailscaleNode(context.TODO(), tt.ip)
			assert.Equal(t, tt.expected, isNode)
		})
	}
}

func TestGetHost_FQDN(t *testing.T) {
	m := new(in.MockClient)
	m.On("QueryDNS", context.TODO(), "test.example.ts.net", "A").Return(
		in.GetTestDNSMessage(), []*dnstype.Resolver{}, nil)
	m.On("WhoIs", context.TODO(), in.TEST_IP.String()).Return(
		&apitype.WhoIsResponse{
			Node: in.GetTestNode([]string{in.TEST_HOST_KEY})},
		nil)
	c := &TSClient{
		Client:  m,
		Tailnet: in.TEST_TAILNET,
	}

	host, err := c.GetHost(context.TODO(), "test.example.ts.net")
	m.AssertExpectations(t)
	require.NoError(t, err)
	assert.NotNil(t, host)
	assert.Equal(t, "test.example.ts.net", host.Name)
	assert.Equal(t, in.TEST_IP, host.IP)
	assert.Len(t, host.Keys, 1)
}

func TestGetHost_Hostname(t *testing.T) {
	m := new(in.MockClient)
	m.On("QueryDNS", context.TODO(), "test.example.ts.net", "A").Return(
		in.GetTestDNSMessage(), []*dnstype.Resolver{}, nil)
	m.On("WhoIs", context.TODO(), in.TEST_IP.String()).Return(
		&apitype.WhoIsResponse{
			Node: in.GetTestNode([]string{in.TEST_HOST_KEY})},
		nil)
	c := &TSClient{
		Client:  m,
		Tailnet: in.TEST_TAILNET,
	}

	host, err := c.GetHost(context.TODO(), "test")
	m.AssertExpectations(t)
	require.NoError(t, err)
	assert.NotNil(t, host)
	assert.Equal(t, "test.example.ts.net", host.Name)
	assert.Equal(t, in.TEST_IP, host.IP)
	assert.Len(t, host.Keys, 1)
}

func TestGetHost_IP(t *testing.T) {
	m := new(in.MockClient)
	m.On("QueryDNS", context.TODO(), in.TEST_IP.String(), "A").Return(
		in.GetTestDNSMessage(), []*dnstype.Resolver{}, nil)
	m.On("WhoIs", context.TODO(), in.TEST_IP.String()).Return(
		&apitype.WhoIsResponse{
			Node: in.GetTestNode([]string{in.TEST_HOST_KEY})},
		nil)
	c := &TSClient{
		Client:  m,
		Tailnet: in.TEST_TAILNET,
	}

	host, err := c.GetHost(context.TODO(), in.TEST_IP.String())
	require.NoError(t, err)
	assert.NotNil(t, host)
	assert.Equal(t, "test.example.ts.net", host.Name, "Expected host name to be resolved from IP")
	assert.Equal(t, in.TEST_IP, host.IP)
	assert.Len(t, host.Keys, 1)
}
