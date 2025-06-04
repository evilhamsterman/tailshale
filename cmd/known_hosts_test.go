package cmd

import (
	"testing"

	in "github.com/evilhamsterman/tailshale/internal"
	ts "github.com/evilhamsterman/tailshale/tailscale"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/ssh"
)

var h = &ts.TailscaleHost{
	Name: "test.example.ts.net",
	IP:   in.TEST_IP,
	Keys: map[string]ssh.PublicKey{
		ts.ED25519: in.TEST_HOST_KEY_OBJECT,
	},
}

func TestGetHostNames(t *testing.T) {
	hosts := getHostNames(h)
	assert.Len(t, hosts, 4)
	assert.Contains(t, hosts, "test.example.ts.net")
	assert.Contains(t, hosts, "test.example.ts.net.")
	assert.Contains(t, hosts, in.TEST_IP.String())
	assert.Contains(t, hosts, "test")
}
