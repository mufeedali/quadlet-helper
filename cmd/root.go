package cmd

import (
	"fmt"
	"os"

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

	// Add subcommands
	rootCmd.AddCommand(backup.BackupCmd)
	rootCmd.AddCommand(cloudflare.CloudflareCmd)
	rootCmd.AddCommand(generate.GenerateCmd)
	rootCmd.AddCommand(unit.UnitCmd)
}

func initConfig() {
	viper.SetEnvPrefix("QH")
	viper.AutomaticEnv()
}
