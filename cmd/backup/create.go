package backup

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/mufeedali/quadlet-helper/internal/backup"
	"github.com/mufeedali/quadlet-helper/internal/shared"
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create [backup-name]",
	Short: "Create a new backup configuration",
	Long:  `Interactive wizard to create a new backup configuration for rsync, restic, or rclone.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(shared.TitleStyle.Render("Create New Backup Configuration"))
		fmt.Println()

		reader := bufio.NewReader(os.Stdin)
		config := &backup.Config{}

		// Get backup name
		if len(args) > 0 {
			config.Name = args[0]
		} else {
			fmt.Print("Backup name: ")
			name, _ := reader.ReadString('\n')
			config.Name = strings.TrimSpace(name)
		}

		if config.Name == "" {
			fmt.Println(shared.ErrorStyle.Render("Error: Backup name is required"))
			os.Exit(1)
		}

		// Check if backup already exists
		configPath, _ := backup.GetConfigPath(config.Name)
		if _, err := os.Stat(configPath); err == nil {
			fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("Error: Backup '%s' already exists", config.Name)))
			os.Exit(1)
		}

		// Get backup type
		fmt.Println("\nBackup type:")
		fmt.Println("  1) rsync  - Local/remote rsync backups")
		fmt.Println("  2) restic - Encrypted incremental backups")
		fmt.Println("  3) rclone - Cloud storage backups")
		fmt.Print("Choose type (1-3): ")
		typeChoice, _ := reader.ReadString('\n')
		typeChoice = strings.TrimSpace(typeChoice)

		switch typeChoice {
		case "1":
			config.Type = backup.BackupTypeRsync
		case "2":
			config.Type = backup.BackupTypeRestic
		case "3":
			config.Type = backup.BackupTypeRclone
		default:
			fmt.Println(shared.ErrorStyle.Render("Invalid choice"))
			os.Exit(1)
		}

		// Check if the selected backup tool is available
		available, err := backup.CheckToolAvailable(config.Type)
		if err != nil {
			fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("Error checking tool availability: %v", err)))
			os.Exit(1)
		}
		if !available {
			fmt.Println()
			fmt.Println(shared.WarningStyle.Render(fmt.Sprintf("⚠ Warning: %s is not installed!", config.Type)))
			fmt.Println()
			fmt.Println(backup.GetInstallInstructions(config.Type))
			fmt.Println()
			if !askYesNo(reader, "Continue anyway? (y/n): ") {
				fmt.Println("Backup creation cancelled.")
				os.Exit(0)
			}
		}

		// Get source paths
		fmt.Println("\nSource paths (comma-separated):")
		fmt.Print("Sources: ")
		sources, _ := reader.ReadString('\n')
		config.Source = strings.Split(strings.TrimSpace(sources), ",")
		for i, s := range config.Source {
			config.Source[i] = strings.TrimSpace(s)
		}

		// Get destination based on type
		fmt.Println("\nDestination:")
		switch config.Type {
		case backup.BackupTypeRsync:
			fmt.Print("Destination path (local or user@host:/path): ")
			dest, _ := reader.ReadString('\n')
			config.Destination.Path = strings.TrimSpace(dest)
		case backup.BackupTypeRestic:
			fmt.Print("Repository path (local or s3:bucket/path, sftp:user@host:/path): ")
			dest, _ := reader.ReadString('\n')
			config.Destination.Repository = strings.TrimSpace(dest)
			fmt.Print("Password file path: ")
			pwFile, _ := reader.ReadString('\n')
			config.Options.PasswordFile = strings.TrimSpace(pwFile)
		case backup.BackupTypeRclone:
			fmt.Print("Remote (e.g., gdrive:backups, s3:bucket/path): ")
			dest, _ := reader.ReadString('\n')
			config.Destination.Remote = strings.TrimSpace(dest)
		}

		// Get schedule
		fmt.Println("\nSchedule:")
		fmt.Println("  Examples: 'daily', 'daily 02:00', 'weekly', 'weekly sun 03:00'")
		fmt.Print("Schedule: ")
		schedule, _ := reader.ReadString('\n')
		config.Schedule = strings.TrimSpace(schedule)

		// Basic options
		fmt.Println("\nOptions:")
		switch config.Type {
		case backup.BackupTypeRsync:
			config.Options.Archive = askYesNo(reader, "Use archive mode (-a)? (y/n): ")
			config.Options.Compress = askYesNo(reader, "Use compression (-z)? (y/n): ")
			config.Options.Delete = askYesNo(reader, "Delete extraneous files (--delete)? (y/n): ")
		case backup.BackupTypeRclone:
			fmt.Print("Number of transfers (default 4): ")
			transfers, _ := reader.ReadString('\n')
			transfers = strings.TrimSpace(transfers)
			if transfers != "" {
				if t, err := strconv.Atoi(transfers); err == nil {
					config.Options.Transfers = t
				}
			}
		}

		// Verification
		fmt.Println("\nVerification:")
		config.Verification.Enabled = askYesNo(reader, "Enable verification? (y/n): ")
		if config.Verification.Enabled {
			config.Verification.AutoVerify = askYesNo(reader, "Auto-verify after each backup? (y/n): ")

			if config.Type == backup.BackupTypeRclone {
				fmt.Println("Verification method:")
				fmt.Println("  1) check     - Compare files")
				fmt.Println("  2) size      - Compare sizes")
				fmt.Println("  3) cryptcheck - For encrypted remotes")
				fmt.Print("Choose method (1-3, default 1): ")
				method, _ := reader.ReadString('\n')
				method = strings.TrimSpace(method)
				switch method {
				case "2":
					config.Verification.Method = "size"
				case "3":
					config.Verification.Method = "cryptcheck"
				default:
					config.Verification.Method = "check"
				}
			}
		}

		// Retention
		fmt.Println("\nRetention:")
		switch config.Type {
		case backup.BackupTypeRestic:
			fmt.Print("Keep daily snapshots (0 to disable): ")
			daily, _ := reader.ReadString('\n')
			if d, err := strconv.Atoi(strings.TrimSpace(daily)); err == nil {
				config.Options.KeepDaily = d
			}
			fmt.Print("Keep weekly snapshots (0 to disable): ")
			weekly, _ := reader.ReadString('\n')
			if w, err := strconv.Atoi(strings.TrimSpace(weekly)); err == nil {
				config.Options.KeepWeekly = w
			}
		case backup.BackupTypeRclone:
			fmt.Print("Keep files for days (0 to disable): ")
			days, _ := reader.ReadString('\n')
			if d, err := strconv.Atoi(strings.TrimSpace(days)); err == nil {
				config.Retention.KeepDays = d
			}
		}

		// Email notifications
		fmt.Println("\nEmail Notifications:")
		config.Notifications.Enabled = askYesNo(reader, "Enable email notifications? (y/n): ")
		if config.Notifications.Enabled {
			config.Notifications.OnFailure = askYesNo(reader, "Notify on failure? (y/n): ")
			config.Notifications.OnSuccess = askYesNo(reader, "Notify on success? (y/n): ")

			fmt.Print("Email to: ")
			to, _ := reader.ReadString('\n')
			config.Notifications.Email.To = strings.TrimSpace(to)

			fmt.Print("Email from: ")
			from, _ := reader.ReadString('\n')
			config.Notifications.Email.From = strings.TrimSpace(from)

			fmt.Print("SMTP host: ")
			host, _ := reader.ReadString('\n')
			config.Notifications.Email.SMTP.Host = strings.TrimSpace(host)

			fmt.Print("SMTP port (default 587): ")
			portStr, _ := reader.ReadString('\n')
			portStr = strings.TrimSpace(portStr)
			if portStr == "" {
				config.Notifications.Email.SMTP.Port = 587
			} else {
				if p, err := strconv.Atoi(portStr); err == nil {
					config.Notifications.Email.SMTP.Port = p
				}
			}

			fmt.Print("SMTP username: ")
			username, _ := reader.ReadString('\n')
			config.Notifications.Email.SMTP.Username = strings.TrimSpace(username)

			fmt.Print("SMTP password file path: ")
			pwFile, _ := reader.ReadString('\n')
			config.Notifications.Email.SMTP.PasswordFile = strings.TrimSpace(pwFile)

			config.Notifications.Email.SMTP.TLS = askYesNo(reader, "Use TLS? (y/n): ")
		}

		// Validate and save
		if err := config.Validate(); err != nil {
			fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("Validation error: %v", err)))
			os.Exit(1)
		}

		if err := backup.SaveConfig(config); err != nil {
			fmt.Println(shared.ErrorStyle.Render(fmt.Sprintf("Error saving config: %v", err)))
			os.Exit(1)
		}

		configPath, _ = backup.GetConfigPath(config.Name)
		fmt.Println()
		fmt.Println(shared.SuccessStyle.Render("✓ Backup configuration created!"))
		fmt.Println(shared.FilePathStyle.Render(configPath))
		fmt.Println()
		fmt.Println("Next steps:")
		fmt.Printf("  1. Install the backup: %s\n", shared.FilePathStyle.Render(fmt.Sprintf("qh backup install %s", config.Name)))
		fmt.Printf("  2. Test the backup: %s\n", shared.FilePathStyle.Render(fmt.Sprintf("qh backup test %s", config.Name)))
	},
}

func askYesNo(reader *bufio.Reader, prompt string) bool {
	fmt.Print(prompt)
	response, _ := reader.ReadString('\n')
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}
