package generate

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/mufeedali/quadlet-helper/internal/shared"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Generate .env.example files from .env files",
	Long: `This command searches for .env files in the specified containers directory
and creates a corresponding .env.example file for each, stripping the values.`,
	Run: func(c *cobra.Command, args []string) {
		containersPath := viper.GetString("containers-path")
		realContainersPath := shared.ResolveContainersDir(containersPath)

		fmt.Println(shared.TitleStyle.Render("Searching for .env files in " + shared.FilePathStyle.Render(realContainersPath) + "..."))

		foundAny := false
		hasErrors := false
		modifiedAny := false

		err := shared.WalkWithSymlinks(realContainersPath, func(path string, d fs.DirEntry) error {
			if d.IsDir() || d.Name() != ".env" {
				return nil
			}

			foundAny = true
			fmt.Println("  -> Found: " + shared.FilePathStyle.Render(path))
			modified, err := processEnvFile(path)
			if err != nil {
				hasErrors = true
				fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("     Error processing %s: %v", path, err)))
			}
			if modified {
				modifiedAny = true
			}

			return nil
		})

		if err != nil {
			fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("Error walking directory: %v", err)))
			os.Exit(1)
		}

		if !foundAny {
			fmt.Println(shared.WarningStyle.Render("No .env files found in any subdirectories."))
		}

		if hasErrors {
			fmt.Println(shared.ErrorStyle.Render("\nGeneration completed with errors."))
			os.Exit(1)
		}

		fmt.Println(shared.TitleStyle.Render("\nGeneration complete."))

		ExitIfHookMode(c, modifiedAny)
	},
}

func processEnvFile(path string) (bool, error) {
	inFile, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer func() {
		_ = inFile.Close()
	}()

	var builder strings.Builder
	scanner := bufio.NewScanner(inFile)
	for scanner.Scan() {
		line := scanner.Text()
		strippedLine := strings.TrimSpace(line)

		if strippedLine == "" || strings.HasPrefix(strippedLine, "#") {
			builder.WriteString(line + "\n")
			continue
		}

		if strings.Contains(strippedLine, "=") {
			key := strings.Split(strippedLine, "=")[0]
			builder.WriteString(key + "=\n")
		}
	}

	if err := scanner.Err(); err != nil {
		return false, err
	}

	exampleContent := builder.String()
	exampleFilePath := filepath.Join(filepath.Dir(path), ".env.example")

	existingContent, err := os.ReadFile(exampleFilePath)
	if err == nil && string(existingContent) == exampleContent {
		return false, nil
	}

	err = os.WriteFile(exampleFilePath, []byte(exampleContent), 0644)
	if err != nil {
		return false, err
	}

	fmt.Println(shared.SuccessStyle.Render("     Generated: ") + shared.FilePathStyle.Render(exampleFilePath))
	return true, nil
}
