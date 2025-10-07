package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var generateTraefikCmd = &cobra.Command{
	Use:   "traefik",
	Short: "Generate a sanitized traefik.yaml.example",
	Long: `This command reads your traefik.yaml, sanitizes sensitive information
like email addresses and network names, and creates a traefik.yaml.example file.`,
	Run: func(cmd *cobra.Command, args []string) {
		containersDir := viper.GetString("containers-dir")
		realContainersDir := resolveContainersDir(containersDir)

		traefikConfigPath := filepath.Join(realContainersDir, "traefik", "container-config", "traefik", "traefik.yaml")
		exampleConfigPath := filepath.Join(realContainersDir, "traefik", "traefik.yaml.example")

		fmt.Println(titleStyle.Render("Starting generation of Traefik config example..."))

		if _, err := os.Stat(traefikConfigPath); os.IsNotExist(err) {
			fmt.Println(errorStyle.Render(fmt.Sprintf("Error: %s not found!", traefikConfigPath)))
			os.Exit(1)
		}

		fmt.Println("  -> Reading: " + filePathStyle.Render(traefikConfigPath))

		content, err := os.ReadFile(traefikConfigPath)
		if err != nil {
			fmt.Println(errorStyle.Render(fmt.Sprintf("Error reading file: %v", err)))
			os.Exit(1)
		}

		sanitizedContent := sanitizeTraefikConfig(string(content))

		err = os.WriteFile(exampleConfigPath, []byte(sanitizedContent), 0644)
		if err != nil {
			fmt.Println(errorStyle.Render(fmt.Sprintf("Error writing example file: %v", err)))
			os.Exit(1)
		}

		fmt.Println(successStyle.Render("     Generated: ") + filePathStyle.Render(exampleConfigPath))
		fmt.Println(titleStyle.Render("\nGeneration complete."))
	},
}

func sanitizeTraefikConfig(content string) string {
	emailPattern := regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`)
	content = emailPattern.ReplaceAllString(content, "your-email@example.com")

	sensitiveFields := map[string]string{
		"network": "your-shared-network",
	}

	for field, replacement := range sensitiveFields {
		// Using a regex that is careful about YAML structure.
		// It looks for `field:` at the beginning of a line, possibly with spaces,
		// and replaces the value that follows.
		pattern := regexp.MustCompile(fmt.Sprintf(`(?m)^(\s*%s:\s*).+$`, regexp.QuoteMeta(field)))
		content = pattern.ReplaceAllString(content, fmt.Sprintf(`${1}%s`, replacement))
	}

	return content
}

func init() {
	generateCmd.AddCommand(generateTraefikCmd)
}
