package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	ts "github.com/evilhamsterman/tailshale/tailscale"
	"github.com/miekg/dns"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/knownhosts"
	"tailscale.com/client/local"
)

var (
	check        bool
	HostKeyTypes struct {
		rsa     bool
		ecdsa   bool
		ed25519 bool
	}
)

var knownHostsCmd = &cobra.Command{
	Use:   "known-hosts",
	Short: "Print SSH host keys for Tailscale nodes",
	Long: strings.TrimLeft(`
This command retrieves and prints the SSH host keys for Tailscale nodes that
have Tailscale SSH enabled. It prints them out in a format compatible with the
SSH known_hosts file.`, "\n"),
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c, _ := ts.NewTSClient(&local.Client{})
		if check && len(args) > 1 {
			cmd.PrintErrln("Error: --check can only be used with a single host")
			return
		} else if check {
			// Check if the host supports Tailscale SSH
			if !CheckHost(args[0], c) {
				fmt.Fprintf(os.Stderr, "Host %s does not support Tailscale SSH.\n", args[0])
				os.Exit(1)
			}
			os.Exit(0)
		}
		PrintKnownHosts(args, c)
	},
}

func init() {
	rootCmd.AddCommand(knownHostsCmd)
	knownHostsCmd.Flags().SortFlags = false
	knownHostsCmd.Flags().BoolVar(&check, "check", false, "Check if the host supports Tailscale SSH")
	knownHostsCmd.Flags().BoolVar(&HostKeyTypes.rsa, "rsa", true, "Include RSA host keys")
	knownHostsCmd.Flags().BoolVar(&HostKeyTypes.ecdsa, "ecdsa", true, "Include ECDSA host keys")
	knownHostsCmd.Flags().BoolVar(&HostKeyTypes.ed25519, "ed25519", true, "Include Ed25519 host keys")
}

// getHostNames generates the hostnames and IP addresses for the given Tailscale node
func getHostNames(host *ts.TailscaleHost) []string {
	cn := dns.CanonicalName(host.Name)
	fqdn := strings.TrimSuffix(cn, ".")
	hostname := dns.SplitDomainName(fqdn)[0]

	return []string{
		host.IP.String(),
		cn,
		fqdn,
		hostname,
	}
}

// CheckHost checks if the given host supports Tailscale SSH
func CheckHost(host string, tsclient *ts.TSClient) bool {
	tsHost, err := tsclient.GetHost(context.Background(), host)
	if err != nil {
		return false
	}
	if tsHost == nil {
		return false
	}
	if len(tsHost.Keys) == 0 {
		return false
	}
	return true
}

// PrintKnownHosts prints the SSH host keys for the given Tailscale nodes.
func PrintKnownHosts(nodes []string, tsclient *ts.TSClient) {

	known_hosts := []string{}
	for _, node := range nodes {
		tsHost, err := tsclient.GetHost(context.Background(), node)
		if err != nil {
			continue
		}
		if tsHost == nil || len(tsHost.Keys) == 0 {
			continue
		}
		hostnames := getHostNames(tsHost)
		var l string
		for keyType, key := range tsHost.Keys {
			switch {
			case keyType == ts.RSA && HostKeyTypes.rsa:
				l = knownhosts.Line(hostnames, key)
			case keyType == ts.ECDSA && HostKeyTypes.ecdsa:
				l = knownhosts.Line(hostnames, key)
			case keyType == ts.ED25519 && HostKeyTypes.ed25519:
				l = knownhosts.Line(hostnames, key)
			}
			known_hosts = append(known_hosts, l)
		}
	}

	if len(known_hosts) == 0 {
		fmt.Fprintln(os.Stderr, "No Tailscale SSH host keys found.")
		os.Exit(1)
	}

	for _, line := range known_hosts {
		fmt.Println(line)
	}
}
