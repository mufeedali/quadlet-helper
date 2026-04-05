package backup

import (
	"strings"

	"github.com/spf13/cobra"
)

// getBackupNameCompletions returns a ValidArgsFunction that provides backup name completions
func getBackupNameCompletions() func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return backupCompletionFunc(nil)
}

// getInstalledBackupCompletions returns completions for backups that are installed
func getInstalledBackupCompletions() func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return backupCompletionFunc(isInstalledBackup)
}

// getNotInstalledBackupCompletions returns completions for backups that are not installed
func getNotInstalledBackupCompletions() func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return backupCompletionFunc(func(name string) bool { return !isInstalledBackup(name) })
}

// getNotifyCompletions returns a ValidArgsFunction that provides completions for the notify command
func getNotifyCompletions() func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		// First argument: backup name
		if len(args) == 0 {
			return backupCompletionFunc(nil)(cmd, args, toComplete)
		}

		// Second argument: status (success or failure)
		if len(args) == 1 {
			statuses := []string{"success", "failure"}
			completions := make([]string, 0, len(statuses))
			for _, status := range statuses {
				if strings.HasPrefix(status, toComplete) {
					completions = append(completions, status)
				}
			}
			return completions, cobra.ShellCompDirectiveNoFileComp
		}

		return nil, cobra.ShellCompDirectiveNoFileComp
	}
}
