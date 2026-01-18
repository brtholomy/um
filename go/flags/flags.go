package flags

import (
	"fmt"
	"reflect"
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

// print out help string by reflecting over fields of provided opts struct
func Help(subcmd string, opts any) {
	v := reflect.ValueOf(opts)
	t := v.Type()

	fmt.Printf("um %s usage:\n", subcmd)
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)
		switch value.Interface().(type) {
		case Arg:
			fmt.Printf("%-12v string\t: %v\n", "["+strings.ToLower(field.Name)+"]", value.FieldByName("Help"))
		case Flag[string]:
			fmt.Printf("%-12v | %v string\t: %v\n", value.FieldByName("Long"), value.FieldByName("Short"), value.FieldByName("Help"))
		case Flag[bool]:
			fmt.Printf("%-12v | %v bool\t: %v\n", value.FieldByName("Long"), value.FieldByName("Short"), value.FieldByName("Help"))
		}
	}
}
