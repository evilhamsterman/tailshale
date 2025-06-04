package internal

import (
	"context"
	"net/netip"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/ssh"
	"tailscale.com/client/tailscale/apitype"
	"tailscale.com/ipn/ipnstate"
	"tailscale.com/tailcfg"
	"tailscale.com/types/dnstype"
)

var (
	TEST_HOST_KEY                    = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAILiup8poNplQGlzXuLDbn2Tz+/L3WxAwimSq7e+eTKjp testkey"
	TEST_TAILNET                     = "example.ts.net"
	TEST_HOST_KEY_OBJECT, _, _, _, _ = ssh.ParseAuthorizedKey([]byte(TEST_HOST_KEY))
	TEST_IP                          = netip.MustParseAddr("100.100.100.100")
)

func GetTestDNSMessage() []byte {
	d := new(dns.Msg)
	d.SetQuestion("test.example.ts.net.", dns.TypeA)
	r, _ := dns.NewRR("test.example.ts.net. IN A 100.100.100.100")
	d.Answer = []dns.RR{r}
	msg, _ := d.Pack()
	return msg
}

func GetTestNode(sshKey []string) *tailcfg.Node {
	h := tailcfg.Hostinfo{
		SSH_HostKeys: sshKey,
	}
	hv := h.View()
	node := &tailcfg.Node{
		Name:     "test." + TEST_TAILNET,
		Hostinfo: hv,
	}
	return node
}

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
