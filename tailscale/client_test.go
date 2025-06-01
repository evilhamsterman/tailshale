package tailscale

import (
	"context"
	"net/netip"
	"testing"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
	"tailscale.com/client/tailscale/apitype"
	"tailscale.com/ipn/ipnstate"
	"tailscale.com/tailcfg"
	"tailscale.com/types/dnstype"
)

const (
	TEST_HOST_KEY = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAILiup8poNplQGlzXuLDbn2Tz+/L3WxAwimSq7e+eTKjp testkey"
	TEST_TAILNET  = "example.ts.net"
)

var (
	TEST_HOST_KEY_OBJECT, _, _, _, _ = ssh.ParseAuthorizedKey([]byte(TEST_HOST_KEY))
	TEST_IP                          = netip.MustParseAddr("100.100.100.100")
)

type MockClient struct {
	mock.Mock
}

func (m *MockClient) Status(ctx context.Context) (*ipnstate.Status, error) {
	args := m.Called(ctx)
	return args.Get(0).(*ipnstate.Status), args.Error(1)
}

func (m *MockClient) QueryDNS(ctx context.Context, host string, qtype string) ([]byte, []*dnstype.Resolver, error) {
	args := m.Called(ctx, host, qtype)
	return args.Get(0).([]byte), args.Get(1).([]*dnstype.Resolver), args.Error(2)
}

func (m *MockClient) WhoIs(ctx context.Context, ip string) (*apitype.WhoIsResponse, error) {
	args := m.Called(ctx, ip)
	return args.Get(0).(*apitype.WhoIsResponse), args.Error(1)
}

var _ Client = (*MockClient)(nil) // Ensure MockClient implements the Client interface

func TestNewTSClient(t *testing.T) {
	m := new(MockClient)
	m.On("Status", mock.Anything).Return(&ipnstate.Status{
		CurrentTailnet: &ipnstate.TailnetStatus{
			MagicDNSSuffix: TEST_TAILNET,
		},
	}, nil)

	client, err := NewTSClient(m)
	m.AssertExpectations(t)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, TEST_TAILNET, client.Tailnet)
}

func TestQueryDNS(t *testing.T) {
	d := new(dns.Msg)
	d.SetQuestion("test.example.ts.net.", dns.TypeA)
	r, _ := dns.NewRR("test.example.ts.net. IN A 100.100.100.100")
	d.Answer = []dns.RR{r}
	msg, _ := d.Pack()

	m := new(MockClient)
	m.On("QueryDNS", context.TODO(), "test.example.ts.net", "A").Return(
		msg, []*dnstype.Resolver{}, nil)

	c := &TSClient{
		Client:  m,
		Tailnet: TEST_TAILNET,
	}
	ip, err := c.QueryTSDNS(context.TODO(), "test.example.ts.net")
	m.AssertExpectations(t)
	assert.NoError(t, err)
	assert.NotNil(t, ip)
	assert.Equal(t, netip.MustParseAddr("100.100.100.100"), *ip)
}

func TestGetSSHHostKeys(t *testing.T) {
	h := tailcfg.Hostinfo{
		SSH_HostKeys: []string{TEST_HOST_KEY},
	}
	hv := h.View()
	m := new(MockClient)
	m.On("WhoIs", context.TODO(), TEST_IP.String()).Return(&apitype.WhoIsResponse{
		Node: &tailcfg.Node{
			Hostinfo: hv,
		},
	}, nil)

	c := &TSClient{
		Client:  m,
		Tailnet: TEST_TAILNET,
	}
	keys, err := c.GetSSHHostKeys(context.TODO(), TEST_IP)
	m.AssertExpectations(t)
	assert.NoError(t, err)
	require.Len(t, keys, 1)
	assert.Equal(t, TEST_HOST_KEY_OBJECT, keys[0])
}

func TestGetSSHHostKeys_NoSSH(t *testing.T) {
	h := tailcfg.Hostinfo{
		SSH_HostKeys: []string{},
	}
	hv := h.View()
	m := new(MockClient)
	m.On("WhoIs", context.TODO(), TEST_IP.String()).Return(&apitype.WhoIsResponse{
		Node: &tailcfg.Node{
			Hostinfo: hv,
		},
	}, nil)

	c := &TSClient{
		Client:  m,
		Tailnet: TEST_TAILNET,
	}
	keys, err := c.GetSSHHostKeys(context.TODO(), TEST_IP)
	m.AssertExpectations(t)
	assert.Error(t, err)
	assert.Nil(t, keys)
}

func TestIsTailcaleNode(t *testing.T) {
	c := &TSClient{
		Tailnet: TEST_TAILNET,
	}
	tests := []struct {
		ip       netip.Addr
		expected bool
	}{
		{
			ip:       TEST_IP,
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
