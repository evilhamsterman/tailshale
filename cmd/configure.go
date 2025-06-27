package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/evilhamsterman/tailshale/internal"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var clean = false

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Configure SSH client for Tailscale",
	Long: strings.TrimLeft(`
Configure the SSH client to automatically retrieve and validate host keys for
Tailscale nodes with Tailscale SSH enabled. This command sets up the necessary
configurations to allow seamless host key authentication for Tailscale nodes.`,
		"\n"),
	Run: func(cmd *cobra.Command, args []string) {
		fs := afero.NewOsFs()
		sshConfigPath := viper.GetString("ssh_config")
		tailshaleCommand, err := os.Executable()
		if err != nil {
			cmd.Println("Error getting executable path:", err)
			os.Exit(1)
		}

		if clean {
			if err := CleanSSHConfig(fs, sshConfigPath); err != nil {
				cmd.Println("Error cleaning SSH configuration:", err)
				os.Exit(1)
			}
			cmd.Println("SSH configuration cleaned")
		} else {
			if err := AddTailshaleConfig(fs, sshConfigPath, tailshaleCommand); err != nil {
				cmd.Println("Error adding include line to SSH config:", err)
				os.Exit(1)
			}

			cmd.Println("Configuration complete. Your SSH client is now set up for Tailscale.")
		}
	},
}

func init() {
	configureCmd.Flags().BoolVar(&clean, "clean", false, "Clean up the SSH configuration by removing the include line and the include file")
	rootCmd.AddCommand(configureCmd)
}

// AddTailshaleConfig adds the include line to the SSH config file
func AddTailshaleConfig(fs afero.Fs, sshConfPath, exePath string) error {
	// Open the file, create it if it doesn't exist
	sshConfFile, err := fs.OpenFile(sshConfPath, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return fmt.Errorf("Error opening ssh config file: %w", err)
	}
	defer sshConfFile.Close()
	cfg, err := internal.NewSSHConfigFromFile(sshConfFile)
	if err != nil {
		return fmt.Errorf("Error reading ssh config file: %w", err)
	}

	cfg.SetConfig(exePath)

	err = sshConfFile.Truncate(0) // Clear the file before writing
	if err != nil {
		return fmt.Errorf("Error truncating ssh config file: %w", err)
	}
	sshConfFile.Seek(0, 0) // Reset the file pointer to the beginning
	_, err = sshConfFile.WriteString(cfg.String())
	if err != nil {
		return fmt.Errorf("Error writing to ssh config file: %w", err)
	}

	return nil
}

// CleanSSHConfig cleans up the SSH configuration by removing the include line and the include file
func CleanSSHConfig(fs afero.Fs, sshConfPath string) error {
	// If the file doesn't exist, there's nothing to clean
	exists, err := afero.Exists(fs, sshConfPath)
	if err != nil {
		return fmt.Errorf("Error checking if ssh config file exists: %w", err)
	}
	if !exists {
		return nil
	}

	// Read the file content
	sshConfFile, err := fs.OpenFile(sshConfPath, os.O_RDWR, 0600)
	if err != nil {
		return fmt.Errorf("Error opening ssh config file: %w", err)
	}
	defer sshConfFile.Close()

	cfg, err := internal.NewSSHConfigFromFile(sshConfFile)
	if err != nil {
		return fmt.Errorf("Error reading ssh config file: %w", err)
	}

	cfg.Config = "" // Clear the config part

	// Write the cleaned config back to the file
	err = sshConfFile.Truncate(0) // Clear the file before writing
	if err != nil {
		return fmt.Errorf("Error truncating ssh config file: %w", err)
	}
	sshConfFile.Seek(0, 0) // Reset the file pointer to the beginning
	_, err = sshConfFile.WriteString(cfg.String())
	if err != nil {
		return fmt.Errorf("Error writing to ssh config file: %w", err)
	}

	return nil
}
