package unit

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var enableCmd = &cobra.Command{
	Use:               "enable <unit-name>...",
	Short:             "Enable one or more quadlet units to start on boot",
	Args:              cobra.MinimumNArgs(1),
	ValidArgsFunction: unitCompletionFunc,
	RunE: func(cmd *cobra.Command, args []string) error {
		return updateInstallSections(args, "enable", func(unitName string) string {
			return fmt.Sprintf("Unit %s is already enabled.", unitName)
		}, addInstallSection, "Added [Install] section to %s")
	},
}

func addInstallSection(content string) (string, bool) {
	if strings.Contains(content, "[Install]") {
		return content, false
	}
	installSection := `
[Install]
WantedBy=multi-user.target default.target
`
	return content + installSection, true
}
