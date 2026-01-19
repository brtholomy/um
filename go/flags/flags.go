package flags

import (
	"fmt"
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

func HasDashPrefix(s string) bool {
	return strings.HasPrefix(s, "-")
}

func MissingValueArg(args []string, i int) bool {
	return i+1 == len(args) || HasDashPrefix(args[i+1])
}

func HelpMissingVal(flag string) {
	fmt.Printf("%s needs a value assignment\n", flag)
}

// print out help string by reflecting over fields of provided opts struct
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
	fmt.Printf("um %s%s\n", subcmd, positional)
	w.Flush()
}
