package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var generateEnvCmd = &cobra.Command{
	Use:   "env",
	Short: "Generate .env.example files from .env files",
	Long: `This command searches for .env files in the specified containers directory
and creates a corresponding .env.example file for each, stripping the values.`,
	Run: func(cmd *cobra.Command, args []string) {
		containersDir := viper.GetString("containers-dir")
		realContainersDir := resolveContainersDir(containersDir)

		fmt.Println(titleStyle.Render("Searching for .env files in " + filePathStyle.Render(realContainersDir) + "..."))

		foundAny := false
		hasErrors := false

		err := walkWithSymlinks(realContainersDir, func(path string, info os.FileInfo) error {
			if !info.IsDir() && info.Name() == ".env" {
				foundAny = true
				fmt.Println("  -> Found: " + filePathStyle.Render(path))
				err := processEnvFile(path)
				if err != nil {
					hasErrors = true
					fmt.Println(errorStyle.Render(fmt.Sprintf("     Error processing %s: %v", path, err)))
				}
			}
			return nil
		})

		if err != nil {
			fmt.Println(errorStyle.Render(fmt.Sprintf("Error walking directory: %v", err)))
			os.Exit(1)
		}

		if !foundAny {
			fmt.Println(warningStyle.Render("No .env files found in any subdirectories."))
		}

		if hasErrors {
			fmt.Println(errorStyle.Render("\nGeneration completed with errors."))
			os.Exit(1)
		}

		fmt.Println(titleStyle.Render("\nGeneration complete."))
	},
}

func processEnvFile(path string) error {
	inFile, err := os.Open(path)
	if err != nil {
		return err
	}
	defer inFile.Close()

	exampleFilePath := filepath.Join(filepath.Dir(path), ".env.example")
	outFile, err := os.Create(exampleFilePath)
	if err != nil {
		return err
	}
	defer outFile.Close()

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

	fmt.Println(successStyle.Render("     Generated: ") + filePathStyle.Render(exampleFilePath))
	return nil
}

func init() {
	generateCmd.AddCommand(generateEnvCmd)
}
