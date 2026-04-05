package cmdutil

import (
	"fmt"

	"github.com/mufeedali/quadlet-helper/internal/shared"
)

func Errorf(format string, args ...any) error {
	return fmt.Errorf(format, args...)
}

func Wrap(err error, format string, args ...any) error {
	if err == nil {
		return nil
	}
	args = append(args, err)
	return fmt.Errorf(format+": %w", args...)
}

func PrintError(err error) {
	if err == nil {
		return
	}
	fmt.Println(shared.ErrorStyle.Render(err.Error()))
}
