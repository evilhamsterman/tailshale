package main

import (
	"context"
	"fmt"

	tsclient "tailscale.com/client/local"
)

func main() {
	ctx := context.TODO()
	s, err := tsclient.Status(ctx)
	if err != nil {
		fmt.Println(err)
	}
	for _, v := range s.Peer {
		if v.HostName == "mysql-qt-events-prod" {
			for _, k := range v.SSH_HostKeys {
				fmt.Println(k)
			}
		}
	}
}
