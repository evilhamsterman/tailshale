package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var Version = "devel"

var rootCmd = &cobra.Command{
	Use:   "tailshale",
	Short: "Automatic hostkey validation for Tailscale",
	Long: `Retrieve hostkeys for Tailscale nodes with Tailscale SSH enabled
	
Can Integrate with the SSH client to allow for seamless hostkey authentication`,
	Version: Version,
}

func Execute() {
	colorScheme := fang.WithColorSchemeFunc(fang.AnsiColorScheme)

	if err := fang.Execute(context.Background(), rootCmd, colorScheme); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().String("config", "", "Configuration file")
	viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))
	rootCmd.PersistentFlags().String("ssh-config", "", "Path to the SSH configuration file")
	viper.BindPFlag("ssh_config", rootCmd.PersistentFlags().Lookup("ssh-config"))
}

func initConfig() {
	confDir, err := os.UserConfigDir()
	if err != nil {
		fmt.Println("Error getting user config directory:", err)
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error getting user home directory:", err)
	}
	// Set defaults
	viper.SetDefault("config", filepath.Join(confDir, "tailshale", "config.yaml"))
	viper.SetDefault("ssh_config", filepath.Join(homeDir, ".ssh/config"))

	// Set the configuration file name and path
	viper.SetEnvPrefix("TAILSHALE")
	viper.AutomaticEnv()
	if configFile := viper.GetString("config"); configFile != "" {
		viper.SetConfigFile(configFile)
	}
	viper.SetConfigName("config")
	viper.AddConfigPath("/etc/tailshale")
	viper.AddConfigPath(".")

	// Read the configuration file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error
		} else {
			// Config file was found but another error was produced
			fmt.Println("Error reading config file:", err)
		}
	}
}
