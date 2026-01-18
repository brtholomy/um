package flags

import (
	"fmt"
	"strings"
)

type Arg struct {
	Val  string
	Help string
}

type Flag struct {
	Long  string
	Short string
	Val   string
	Help  string
}

// there's got to be a way to generalize the val field to string|bool
type FlagBool struct {
	Long  string
	Short string
	Val   bool
	Help  string
}

func HasDashPrefix(s string) bool {
	return strings.HasPrefix(s, "-")
}

func MissingValueArg(args []string, i int) bool {
	return i+1 == len(args) || HasDashPrefix(args[i+1])
}

func HelpMissingVal(flag string) {
	fmt.Printf("%s needs a value assignment\n", flag)
}
