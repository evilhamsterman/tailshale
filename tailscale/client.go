package tailscale

import (
	"context"

	"tailscale.com/client/local"
)

type TSClient struct {
	*local.Client
	tailnet string
}

func NewTSClient() (*TSClient, error) {
	client := &TSClient{
		Client: &local.Client{},
	}
	ctx := context.Background()
	status, err := client.Client.Status(ctx)
	if err != nil {
		return nil, err
	}
	client.tailnet = status.CurrentTailnet.MagicDNSSuffix
	return client, nil
}
