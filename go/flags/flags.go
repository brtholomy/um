package flags

import (
	"fmt"
	"strings"
)

type Arg struct {
	Val  string
	Help string
}

// type constraint
type Val interface {
	string | bool
}

// with generic Val field, instantiate like Flag[string]
type Flag[V Val] struct {
	Long  string
	Short string
	Val   V
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
