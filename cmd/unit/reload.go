package unit

import (
	"fmt"

	"github.com/mufeedali/quadlet-helper/internal/cmdutil"
	"github.com/mufeedali/quadlet-helper/internal/shared"
	"github.com/mufeedali/quadlet-helper/internal/systemd"
)

// reloadDaemon performs a systemctl daemon-reload and prints UI messages.
// Callers should decide whether to call it (i.e., respect --no-reload) so the
// function itself does not print skip messages.
func reloadDaemon() error {
	fmt.Println(shared.TitleStyle.Render("Reloading systemctl daemon..."))
	output, err := systemd.DaemonReload()
	if err != nil {
		fmt.Println(shared.InfoMark + " Please reload manually: systemctl --user daemon-reload")
		if output != "" {
			return cmdutil.Errorf("reloading systemctl daemon: %v\n%s", err, output)
		}
		return cmdutil.Wrap(err, "reloading systemctl daemon")
	}
	if len(output) > 0 {
		fmt.Println(string(output))
	}
	fmt.Println(shared.SuccessStyle.Render("✓ systemctl daemon reloaded"))
	return nil
}

// printSkipReload prints the message shown when --no-reload is used.
func printSkipReload() {
	fmt.Println(shared.InfoMark + " Skipping systemctl daemon reload ( --no-reload )")
}
