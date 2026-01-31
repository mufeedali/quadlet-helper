package generate

import (
	"os"

	"github.com/spf13/cobra"
)

// ExitIfHookMode exits with 1 if `--hook` is set and `modified` is true,
// with 0 if `--hook` is set and `modified` is false. No-op otherwise.
func ExitIfHookMode(c *cobra.Command, modified bool) {
	// Search for the `hook` flag across likely flag sets (local flags, persistent flags
	// on command, parent, and root) to handle different registration locations.
	var (
		hookMode bool
		found    bool
	)

	// Order: command, parent, root
	cmds := []*cobra.Command{c}
	if c.Parent() != nil {
		cmds = append(cmds, c.Parent())
	}
	if c.Root() != nil {
		cmds = append(cmds, c.Root())
	}

	for _, cmd := range cmds {
		if cmd == nil {
			continue
		}

		// Try local flags
		if fs := cmd.Flags(); fs != nil {
			if v, err := fs.GetBool("hook"); err == nil {
				hookMode = v
				found = true
				break
			}
		}

		// Try persistent flags
		if pfs := cmd.PersistentFlags(); pfs != nil {
			if v, err := pfs.GetBool("hook"); err == nil {
				hookMode = v
				found = true
				break
			}
		}
	}

	// If we couldn't find the flag or it's not set, do nothing.
	if !found || !hookMode {
		return
	}

	if modified {
		os.Exit(1)
	}
	os.Exit(0)
}
