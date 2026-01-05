package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mufeedali/quadlet-helper/cmd/backup"
	"github.com/mufeedali/quadlet-helper/cmd/cloudflare"
	"github.com/mufeedali/quadlet-helper/cmd/generate"
	"github.com/mufeedali/quadlet-helper/cmd/unit"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "qh",
	Short: "A helper for quadlet containers",
	Long: `qh is a command-line tool to help manage quadlet-based container setups.
It provides utilities for generating example files and managing services like the Cloudflare IP updater.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error finding home directory:", err)
		os.Exit(1)
	}
	defaultContainersDir := fmt.Sprintf("%s/.config/containers/systemd", home)

	rootCmd.PersistentFlags().String("containers-path", defaultContainersDir, "Root directory for container configurations")
	if err := viper.BindPFlag("containers-path", rootCmd.PersistentFlags().Lookup("containers-path")); err != nil {
		fmt.Println("Error binding flag:", err)
		os.Exit(1)
	}

	// Email configuration
	rootCmd.PersistentFlags().String("email-host", "", "SMTP host for email notifications")
	rootCmd.PersistentFlags().Int("email-port", 587, "SMTP port for email notifications")
	rootCmd.PersistentFlags().String("email-user", "", "SMTP username for email notifications")
	rootCmd.PersistentFlags().String("email-password-file", "", "Path to file containing SMTP password")
	rootCmd.PersistentFlags().Bool("email-tls", true, "Use TLS for SMTP connection")
	rootCmd.PersistentFlags().String("email-from", "", "Default from address for email notifications")
	rootCmd.PersistentFlags().String("email-to", "", "Default to address for email notifications")

	_ = viper.BindPFlag("email.host", rootCmd.PersistentFlags().Lookup("email-host"))
	_ = viper.BindPFlag("email.port", rootCmd.PersistentFlags().Lookup("email-port"))
	_ = viper.BindPFlag("email.user", rootCmd.PersistentFlags().Lookup("email-user"))
	_ = viper.BindPFlag("email.passwordfile", rootCmd.PersistentFlags().Lookup("email-password-file"))
	_ = viper.BindPFlag("email.tls", rootCmd.PersistentFlags().Lookup("email-tls"))
	_ = viper.BindPFlag("email.from", rootCmd.PersistentFlags().Lookup("email-from"))
	_ = viper.BindPFlag("email.to", rootCmd.PersistentFlags().Lookup("email-to"))

	// Add subcommands
	rootCmd.AddCommand(backup.BackupCmd)
	rootCmd.AddCommand(cloudflare.CloudflareCmd)
	rootCmd.AddCommand(generate.GenerateCmd)
	rootCmd.AddCommand(unit.UnitCmd)
}

func initConfig() {
	// Set up config file using XDG Base Directory specification
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Printf("Error finding home directory: %v\n", err)
		} else {
			configHome = filepath.Join(home, ".config")
		}
	}

	if configHome != "" {
		configPath := filepath.Join(configHome, "quadlet-helper")
		viper.AddConfigPath(configPath)
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	// Read config file if it exists
	if err := viper.ReadInConfig(); err != nil {
		// Config file not found is okay, we'll use flags/env vars
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Config file was found but another error occurred
			fmt.Printf("Error reading config file: %v\n", err)
		}
	}

	viper.SetEnvPrefix("QH")
	viper.AutomaticEnv()
}
