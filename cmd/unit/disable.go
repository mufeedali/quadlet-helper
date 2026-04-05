package unit

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var disableCmd = &cobra.Command{
	Use:               "disable <unit-name>...",
	Short:             "Disable one or more quadlet units from starting on boot",
	Args:              cobra.MinimumNArgs(1),
	ValidArgsFunction: unitCompletionFunc,
	RunE: func(cmd *cobra.Command, args []string) error {
		return updateInstallSections(args, "disable", func(unitName string) string {
			return fmt.Sprintf("Unit %s is not enabled, nothing to do.", unitName)
		}, removeInstallSection, "Removed [Install] section from %s")
	},
}

func removeInstallSection(content string) (string, bool) {
	lines := strings.SplitAfter(content, "\n")
	var builder strings.Builder
	removed := false
	inInstall := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(strings.TrimRight(line, "\r\n"))
		if len(trimmed) > 1 && strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			if trimmed == "[Install]" {
				removed = true
				inInstall = true
				continue
			}
			inInstall = false
		}

		if inInstall {
			continue
		}

		builder.WriteString(line)
	}

	if !removed {
		return content, false
	}

	return builder.String(), true
}
