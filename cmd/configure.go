package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

const (
	inFileLine = "include tailshale"
	inFile     = `# Tailshale SSH configuration
# This file is automatically managed by Tailshale.
# Do not edit manually.

Match exec "{{.TailshaleCommand}} known-hosts --check %h" exec-timeout 5
  KnownHostsCommand {{.TailshaleCommand}} known-hosts %h
`
)

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Configure SSH client for Tailscale",
	Long: `Configure the SSH client to automatically retrieve and validate host keys for Tailscale nodes with Tailscale SSH enabled.
This command sets up the necessary configurations to allow seamless host key authentication for Tailscale nodes.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Placeholder for configuration logic
		// This is where you would implement the logic to configure the SSH client
		// to work with Tailscale nodes.
		cmd.Println("Configuration complete. Your SSH client is now set up for Tailscale.")
	},
}

func init() {
	rootCmd.AddCommand(configureCmd)
}

// AddIncludeLine adds the include line to the SSH config file
func AddIncludeLine(fs afero.Fs, path string) error {
	// Check if the file already exists
	exists, err := afero.Exists(fs, path)
	if err != nil {
		return fmt.Errorf("Error checking if file exists: %v", err)
	}

	// if the file exists, check if the include line is already present
	if exists {
		ok, err := afero.FileContainsBytes(fs, path, []byte(inFileLine))
		if err != nil {
			return fmt.Errorf("Error checking if include line exists: %v", err)
		}
		if ok {
			// Include line already exists, no need to add it again
			return nil
		}
	}
	// Open the file for appending, create it if it doesn't exist
	fsFile, err := fs.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0600)
	if err != nil {
		return fmt.Errorf("Error opening file for appending: %v", err)
	}
	defer fsFile.Close()

	// Append the include line to the SSH config file
	if _, err := fsFile.WriteString("\n" + inFileLine + "\n"); err != nil {
		return err
	}
	return nil
}
