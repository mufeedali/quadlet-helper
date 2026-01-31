package generate

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/mufeedali/quadlet-helper/internal/shared"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var traefikCmd = &cobra.Command{
	Use:   "traefik",
	Short: "Generate a sanitized traefik.yaml.example",
	Long: `This command reads your traefik.yaml, sanitizes sensitive information
like email addresses and network names, and creates a traefik.yaml.example file.`,
	Run: func(c *cobra.Command, args []string) {
		containersPath := viper.GetString("containers-path")
		realContainersPath := shared.ResolveContainersDir(containersPath)

		traefikConfigPath := filepath.Join(realContainersPath, "traefik", "container-config", "traefik", "traefik.yaml")
		exampleConfigPath := filepath.Join(realContainersPath, "traefik", "traefik.yaml.example")

		fmt.Println(shared.TitleStyle.Render("Starting generation of Traefik config example..."))

		if _, err := os.Stat(traefikConfigPath); os.IsNotExist(err) {
			fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("Error: %s not found!", traefikConfigPath)))
			os.Exit(1)
		}

		fmt.Println("  -> Reading: " + shared.FilePathStyle.Render(traefikConfigPath))

		content, err := os.ReadFile(traefikConfigPath)
		if err != nil {
			fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("Error reading file: %v", err)))
			os.Exit(1)
		}

		sanitizedContent := sanitizeTraefikConfig(string(content))

		existing, err := os.ReadFile(exampleConfigPath)
		modified := false
		if err != nil || string(existing) != sanitizedContent {
			err = os.WriteFile(exampleConfigPath, []byte(sanitizedContent), 0644)
			if err != nil {
				fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("Error writing example file: %v", err)))
				os.Exit(1)
			}
			fmt.Println(shared.SuccessStyle.Render("     Generated: ") + shared.FilePathStyle.Render(exampleConfigPath))
			modified = true
		}

		fmt.Println(shared.TitleStyle.Render("\nGeneration complete."))

		ExitIfHookMode(c, modified)
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
