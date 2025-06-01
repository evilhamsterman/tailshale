package tailscale

import (
	"context"
	"net"
	"testing"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"tailscale.com/ipn/ipnstate"
	"tailscale.com/types/dnstype"
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

var _ Client = (*MockClient)(nil) // Ensure MockClient implements the Client interface

func TestNewTSClient(t *testing.T) {
	m := new(MockClient)
	m.On("Status", mock.Anything).Return(&ipnstate.Status{
		CurrentTailnet: &ipnstate.TailnetStatus{
			MagicDNSSuffix: "example.ts.net",
		},
	}, nil)

	client, err := NewTSClient(m)
	m.AssertExpectations(t)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, "example.ts.net", client.Tailnet)
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
		Tailnet: "example.ts.net",
	}
	ip, err := c.QueryTSDNS(context.TODO(), "test.example.ts.net")
	m.AssertExpectations(t)
	assert.NoError(t, err)
	assert.NotNil(t, ip)
	assert.Equal(t, net.ParseIP("100.100.100.100"), *ip)
}
