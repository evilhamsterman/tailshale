package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
		includeFilePath := filepath.Join(filepath.Dir(sshConfigPath), "tailshale")
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
			if err := WriteIncludeFile(fs, includeFilePath, tailshaleCommand); err != nil {
				cmd.Println("Error writing include file:", err)
				os.Exit(1)
			}
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

// WriteIncludeFile writes the file to include
func WriteIncludeFile(fs afero.Fs, path string, tailshaleCommand string) error {
	// Open the file for writing, create it if it doesn't exist
	fsFile, err := fs.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("Error opening file for writing: %v", err)
	}
	defer fsFile.Close()

	// Write the include file content
	t := template.New("includeFileTemplate")
	t, err = t.Parse(inFile)
	if err != nil {
		return fmt.Errorf("Error parsing template: %v", err)
	}
	data := struct {
		TailshaleCommand string
	}{
		TailshaleCommand: tailshaleCommand,
	}

	if err := t.Execute(fsFile, data); err != nil {
		return fmt.Errorf("Error executing template: %v", err)
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
	if exists {
		ok, err := afero.FileContainsBytes(fs, sshConfigPath, []byte(inFileLine))
		if err != nil {
			return fmt.Errorf("Error checking if include line exists: %v", err)
		}
		if ok {
			content, err := afero.ReadFile(fs, sshConfigPath)
			if err != nil {
				return fmt.Errorf("Error reading SSH config file: %v", err)
			}
			// Remove the include line from the content
			newContent := strings.ReplaceAll(string(content), inFileLine+"\n", "")
			newContent = strings.ReplaceAll(newContent, inFileLine, "") // In case it was the last line
			err = afero.WriteFile(fs, sshConfigPath, []byte(newContent), 0600)
			if err != nil {
				return fmt.Errorf("Error writing updated SSH config file: %v", err)
			}
		}
	}

	// Remove the include file
	includeFilePath := filepath.Join(filepath.Dir(sshConfigPath), "tailshale")
	if err := fs.Remove(includeFilePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("Error removing include file: %v", err)
	}

	return nil
}
