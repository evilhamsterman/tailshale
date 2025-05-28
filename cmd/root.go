package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "tailshale",
	Short: "Automatic hostkey validation for Tailscale",
	Long: `Retrieve hostkeys for Tailscale nodes with Tailscale SSH enabled
	
Can Integrate with the SSH client to allow for seamless hostkey authentication`,

	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Hello There")
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
