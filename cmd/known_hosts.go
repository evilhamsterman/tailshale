package cmd

import (
	"strings"

	ts "github.com/evilhamsterman/tailshale/tailscale"
	"github.com/spf13/cobra"
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
		// This command will be implemented in the future
		cmd.Println("This command is not yet implemented")
		cmd.Println(args)
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

// PrintKnownHosts prints the SSH host keys for the given Tailscale nodes.
func PrintKnownHosts(nodes []string, tsclient *ts.TSClient) {
	// This function will be implemented in the future
	// It should retrieve the host keys for the given nodes and print them
	// in a format compatible with the SSH known_hosts file.
	for _, node := range nodes {
		// Placeholder for actual host key retrieval logic
		println("Host:", node)
		println("HostKey: [Placeholder for actual host key]")
	}
}
