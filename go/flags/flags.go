package flags

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
	"text/tabwriter"
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
// NOTE: struct inheritance would be easier, but Flag[bool] is much cooler than FlagBool.
// and actually more readable for the initOpts function, which serves as documentation.
// so the complexity seems to be all paid up front here in this pkg.
type Flag[T Val] struct {
	Long  string
	Short string
	Val   T
	Help  string
}

// generic Flag validity check
// NOTE: you cannot define a method on an instantiated type, only the generic
// https://www.reddit.com/r/golang/comments/1n6xasx/comment/nc3dbhd/
func (f Flag[Val]) IsSet() bool {
	// must do this weird dance: convert to empty interface, then cast:
	switch any(f.Val).(type) {
	case string:
		return any(f.Val).(string) != ""
	case bool:
		return any(f.Val).(bool)
	default:
		return false
	}
}

func (f Arg) IsSet() bool {
	return f.Val != ""
}

func HasDashPrefix(s string) bool {
	return strings.HasPrefix(s, "-")
}

// check for non-dashed value ahead in the args slice
func missingValue(args []string, i int) bool {
	return i+1 == len(args) || HasDashPrefix(args[i+1])
}

// if no value ahead in the args, print error and exit
func validValueOrExit(args []string, i int) {
	if missingValue(args, i) {
		log.Printf("um: %s needs a value assignment\n", args[i])
		log.Fatal("try: um [cmd] --help")
	}
}

// 1. validate a Flag[string] by looking ahead in args, if missing, os.Exit(1)
// 2. increment the i to skip that value and return
// 3. fetch that value and return
func ValidateIncrementFetchOrExit(args []string, i int) (int, string) {
	validValueOrExit(args, i)
	return i + 1, args[i+1]
}

// print out help string by reflecting over fields of provided opts struct
// and os.Exit(0)
func Help(subcmd string, opts any) {
	v := reflect.ValueOf(opts)
	t := v.Type()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	positional := ""

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)
		switch value.Interface().(type) {
		case Arg:
			positional += fmt.Sprintf(" [%s]", strings.ToLower(field.Name))
			fmt.Fprintf(w, "[%v]\tstring\t%v\n", strings.ToLower(field.Name), value.FieldByName("Help"))
		case Flag[string], Flag[bool]:
			if field.Name != "Help" {
				positional += fmt.Sprintf(" [%s]", value.FieldByName("Long"))
			}
			fmt.Fprintf(w, "%v | %v\t%v\t%v\n",
				value.FieldByName("Long"),
				value.FieldByName("Short"),
				value.FieldByName("Val").Type(),
				value.FieldByName("Help"),
			)
		}
	}
	log.Printf("um %s%s\n", subcmd, positional)
	w.Flush()
	os.Exit(0)
}
