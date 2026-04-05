package unit

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/mufeedali/quadlet-helper/internal/cmdutil"
	"github.com/mufeedali/quadlet-helper/internal/shared"
	"github.com/spf13/cobra"
)

var logsCmd = &cobra.Command{
	Use:               "logs <unit-name>...",
	Short:             "View logs of one or more quadlet units",
	Args:              cobra.MinimumNArgs(1),
	ValidArgsFunction: unitCompletionFunc,
	RunE: func(cmd *cobra.Command, args []string) error {
		services, err := loadServices(args)
		if err != nil {
			return err
		}

		follow, _ := cmd.Flags().GetBool("follow")

		fmt.Println(shared.TitleStyle.Render(fmt.Sprintf("Logs for %s...", strings.Join(services, " "))))

		// Build journalctl args: --user followed by multiple -u <service> entries
		argsList := []string{"--user"}
		for _, svc := range services {
			argsList = append(argsList, "-u", svc)
		}
		if follow {
			argsList = append(argsList, "-f")
		}

		c := exec.Command("journalctl", argsList...)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		err = c.Run()
		if err != nil {
			return cmdutil.Wrap(err, "getting logs")
		}

		return nil
	},
}

func init() {
	logsCmd.Flags().BoolP("follow", "f", false, "Follow log output")
}
