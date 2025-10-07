package cmd

import (
	"fmt"
	"os"

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

	rootCmd.PersistentFlags().String("containers-dir", defaultContainersDir, "Root directory for container configurations")
	viper.BindPFlag("containers-dir", rootCmd.PersistentFlags().Lookup("containers-dir"))
}

func initConfig() {
	viper.AutomaticEnv()
}
