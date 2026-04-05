package backup

import (
	"os"
	"os/exec"
	"strings"

	internalbackup "github.com/mufeedali/quadlet-helper/internal/backup"
	"github.com/mufeedali/quadlet-helper/internal/cmdutil"
	"github.com/mufeedali/quadlet-helper/internal/shared"
	"github.com/spf13/cobra"
)

func loadBackupConfig(backupName string) (*internalbackup.Config, error) {
	config, err := internalbackup.LoadConfig(backupName)
	if err != nil {
		return nil, cmdutil.Wrap(err, "loading config")
	}
	return config, nil
}

func backupCompletionFunc(filter func(string) bool) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) > 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		configs, err := internalbackup.ListConfigs()
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		completions := make([]string, 0, len(configs))
		for _, name := range configs {
			if !strings.HasPrefix(name, toComplete) {
				continue
			}
			if filter != nil && !filter(name) {
				continue
			}
			completions = append(completions, name)
		}

		return completions, cobra.ShellCompDirectiveNoFileComp
	}
}

func isInstalledBackup(backupName string) bool {
	timerFilePath, err := internalbackup.GetTimerFilePath(backupName)
	if err != nil {
		return false
	}
	return shared.FileExists(timerFilePath)
}

func runJournalctl(args []string) error {
	cmd := exec.Command("journalctl", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}
