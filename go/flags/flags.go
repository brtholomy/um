package flags

import (
	"fmt"
	"strings"
)

type Flag string

func (f Flag) Dash() string {
	return fmt.Sprintf("--%s", f)
}

// prepend --flag to args if a flag isn't there:
// allows trailing flags:
// um tag foo --invert
// becomes:
// um tag --query foo --invert
func PrependFlagToArgs(args []string, flag Flag) []string {
	if len(args) >= 1 && !strings.HasPrefix(args[0], "-") {
		args = append([]string{flag.Dash()}, args...)
	}
	return args
}
