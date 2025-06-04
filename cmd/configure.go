package cmd

import (
	"fmt"
	"html/template"
	"os"
	"strings"

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

		if clean {
			if err := CleanSSHConfig(fs, sshConfigPath); err != nil {
				cmd.Println("Error cleaning SSH configuration:", err)
				os.Exit(1)
			}
			cmd.Println("SSH configuration cleaned")
		} else {
			if err := AddIncludeLine(fs, sshConfigPath); err != nil {
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

// templateTailshaleLine creates the include line for the SSH config file
func templateTailshaleLine() string {
	tailshaleCommand, err := os.Executable()
	if err != nil {
		panic(fmt.Sprintf("Error getting executable path: %v\n", err))
	}

	t := template.New("knownhostscommand")
	t, _ = t.Parse("KnownHostsCommand {{.TailshaleCommand}} known-hosts %h")

	data := struct {
		TailshaleCommand string
	}{
		TailshaleCommand: tailshaleCommand,
	}

	var buf strings.Builder
	_ = t.Execute(&buf, data)
	return buf.String()
}

// AddIncludeLine adds the include line to the SSH config file
func AddIncludeLine(fs afero.Fs, path string) error {
	// Check if the file already exists
	exists, err := afero.Exists(fs, path)
	if err != nil {
		return fmt.Errorf("Error checking if file exists: %v", err)
	}

	// if the file exists, check if the include line is already present

	line := templateTailshaleLine()
	if exists {
		ok, err := afero.FileContainsBytes(fs, path, []byte(line))
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
	if _, err := fsFile.WriteString("\n" + line + "\n"); err != nil {
		return err
	}
	return nil
}

// CleanSSHConfig cleans up the SSH configuration by removing the include line and the include file
func CleanSSHConfig(fs afero.Fs, sshConfigPath string) error {
	// Remove the include line from the SSH config file
	exists, err := afero.Exists(fs, sshConfigPath)
	if err != nil {
		return fmt.Errorf("Error checking if SSH config file exists: %v", err)
	}
	line := templateTailshaleLine()
	if exists {
		ok, err := afero.FileContainsBytes(fs, sshConfigPath, []byte(line))
		if err != nil {
			return fmt.Errorf("Error checking if include line exists: %v", err)
		}
		if ok {
			content, err := afero.ReadFile(fs, sshConfigPath)
			if err != nil {
				return fmt.Errorf("Error reading SSH config file: %v", err)
			}
			// Remove the include line from the content
			newContent := strings.ReplaceAll(string(content), line+"\n", "")
			newContent = strings.ReplaceAll(newContent, line, "") // In case it was the last line
			err = afero.WriteFile(fs, sshConfigPath, []byte(newContent), 0600)
			if err != nil {
				return fmt.Errorf("Error writing updated SSH config file: %v", err)
			}
		}
	}

	return nil
}
