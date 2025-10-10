package backup

import (
	"github.com/mufeedali/quadlet-helper/internal/backup"
	"github.com/spf13/cobra"
)

// getBackupNameCompletions returns a ValidArgsFunction that provides backup name completions
func getBackupNameCompletions() func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		// Only complete the first argument (backup name)
		if len(args) > 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		configs, err := backup.ListConfigs()
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		return configs, cobra.ShellCompDirectiveNoFileComp
	}
}

// getInstalledBackupCompletions returns completions for backups that are installed
func getInstalledBackupCompletions() func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		// Only complete the first argument (backup name)
		if len(args) > 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		configs, err := backup.ListConfigs()
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		// Filter to only installed backups
		var installed []string
		for _, name := range configs {
			timerFilePath, _ := backup.GetTimerFilePath(name)
			if fileExists(timerFilePath) {
				installed = append(installed, name)
			}
		}

		return installed, cobra.ShellCompDirectiveNoFileComp
	}
}

// getNotInstalledBackupCompletions returns completions for backups that are not installed
func getNotInstalledBackupCompletions() func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		// Only complete the first argument (backup name)
		if len(args) > 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		configs, err := backup.ListConfigs()
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		// Filter to only non-installed backups
		var notInstalled []string
		for _, name := range configs {
			timerFilePath, _ := backup.GetTimerFilePath(name)
			if !fileExists(timerFilePath) {
				notInstalled = append(notInstalled, name)
			}
		}

		return notInstalled, cobra.ShellCompDirectiveNoFileComp
	}
}

// getNotifyCompletions returns a ValidArgsFunction that provides completions for the notify command
func getNotifyCompletions() func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		// First argument: backup name
		if len(args) == 0 {
			configs, err := backup.ListConfigs()
			if err != nil {
				return nil, cobra.ShellCompDirectiveError
			}
			return configs, cobra.ShellCompDirectiveNoFileComp
		}

		// Second argument: status (success or failure)
		if len(args) == 1 {
			return []string{"success", "failure"}, cobra.ShellCompDirectiveNoFileComp
		}

		return nil, cobra.ShellCompDirectiveNoFileComp
	}
}
