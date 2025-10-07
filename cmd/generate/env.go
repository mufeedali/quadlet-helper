package generate

import (
	"bufio"
	"fmt"
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
		containersDir := viper.GetString("containers-dir")
		realContainersDir := shared.ResolveContainersDir(containersDir)

		fmt.Println(shared.TitleStyle.Render("Searching for .env files in " + shared.FilePathStyle.Render(realContainersDir) + "..."))

		foundAny := false
		hasErrors := false

		err := shared.WalkWithSymlinks(realContainersDir, func(path string, info os.FileInfo) error {
			if !info.IsDir() && info.Name() == ".env" {
				foundAny = true
				fmt.Println("  -> Found: " + shared.FilePathStyle.Render(path))
				err := processEnvFile(path)
				if err != nil {
					hasErrors = true
					fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("     Error processing %s: %v", path, err)))
				}
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
	},
}

func processEnvFile(path string) error {
	inFile, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() {
		_ = inFile.Close()
	}()

	exampleFilePath := filepath.Join(filepath.Dir(path), ".env.example")
	outFile, err := os.Create(exampleFilePath)
	if err != nil {
		return err
	}
	defer func() {
		_ = outFile.Close()
	}()

	scanner := bufio.NewScanner(inFile)
	for scanner.Scan() {
		line := scanner.Text()
		strippedLine := strings.TrimSpace(line)

		if strippedLine == "" || strings.HasPrefix(strippedLine, "#") {
			_, err := outFile.WriteString(line + "\n")
			if err != nil {
				return err
			}
			continue
		}

		if strings.Contains(strippedLine, "=") {
			key := strings.Split(strippedLine, "=")[0]
			_, err := outFile.WriteString(key + "=\n")
			if err != nil {
				return err
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	fmt.Println(shared.SuccessStyle.Render("     Generated: ") + shared.FilePathStyle.Render(exampleFilePath))
	return nil
}
