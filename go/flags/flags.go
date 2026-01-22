package flags

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
	"text/tabwriter"

	"github.com/brtholomy/um/go/cmd"
)

const INTERNAL_ERR_PREFIX = "um internal err: ParseArgs"

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
func validValueOrExit(sub cmd.Subcommand, args []string, i int) {
	if missingValue(args, i) {
		log.Printf("um %s: %s needs a value assignment\n", sub, args[i])
		log.Fatal("try: um [cmd] --help")
	}
}

// 1. validate a Flag[string] by looking ahead in args, if missing, os.Exit(1)
// 2. increment the i to skip that value and return
// 3. fetch that value and return
func ValidateIncrementFetchOrExit(sub cmd.Subcommand, args []string, i int) (int, string) {
	validValueOrExit(sub, args, i)
	return i + 1, args[i+1]
}

func HelpRequired(sub cmd.Subcommand, long string) {
	log.Printf("um %s: %s is required", sub, long)
}

func HelpInvalidArg(sub cmd.Subcommand, arg string) {
	log.Printf("um %s: invalid argument: %s", sub, arg)
}

// print out help string by reflecting over fields of provided opts struct
// and os.Exit(0)
func Help(sub cmd.Subcommand, summary string, opts any) {
	v := reflect.ValueOf(opts)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	t := v.Type()

	// everything should go to either stderr or stdout:
	w := tabwriter.NewWriter(os.Stderr, 0, 0, 3, ' ', 0)
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
	log.Printf("um %s%s\n\n", sub, positional)
	log.Printf("%s\n\n", summary)
	w.Flush()
	os.Exit(0)
}

// parse the commandline args
//
// reflect on the incoming opts struct filled out with flags.Arg, flags.Flag[bool], flags.Flag[string]
// and fill their "Val" field appropriately.
//
// WARN: incoming opts must be a pointer!
// WORB: opts struct must list its Arg types first and in expected order
// NOTE: uglier than the pretty switch statements I had per cmd, but much more convenient.
func ParseArgs(sub cmd.Subcommand, summary string, args []string, opts any) {
	// immediately get the underlying value of the incoming pointer:
	v := reflect.ValueOf(opts)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	} else {
		log.Fatalf("%s needs a pointer to a struct: %#v", INTERNAL_ERR_PREFIX, opts)
	}
	if v.Kind() != reflect.Struct {
		log.Fatalf("%s needs an underlying struct: %#v", INTERNAL_ERR_PREFIX, opts)
	}

	// for all args, iterate through all opts fields and reflect on type
argloop:
	for i := 0; i < len(args); i++ {
		arg := args[i]

		for j := 0; j < v.NumField(); j++ {
			field := v.Field(j)
			if field.Kind() != reflect.Struct {
				log.Fatalf("%s needs a struct of flag types: %#v is type: %s", INTERNAL_ERR_PREFIX, field, field.Kind())
			}
			val := field.FieldByName("Val")
			// NORT: FieldByName().String() won't panic if not present, as with Arg type, just prints "<invalid Value>"
			long := field.FieldByName("Long").String()
			short := field.FieldByName("Short").String()
			if !val.CanSet() {
				log.Fatalf("%s failed to set flag value: %#v", INTERNAL_ERR_PREFIX, field)
			}
			switch field.Interface().(type) {
			case Arg:
				// NOTE: we assume the opts struct lists its Arg types first and in order
				if j == i && !HasDashPrefix(arg) {
					val.SetString(arg)
					continue argloop
				}
			case Flag[bool]:
				if arg == long || arg == short {
					val.SetBool(true)
					if long == "--help" {
						Help(sub, summary, opts)
					}
					continue argloop
				}
			case Flag[string]:
				if arg == long || arg == short {
					fetch := ""
					i, fetch = ValidateIncrementFetchOrExit(sub, args, i)
					val.SetString(fetch)
					continue argloop
				}
			default:
				log.Fatalf("um internal err: unrecognized flag type in opts struct: %#v", field)
			}
		}
		// BLAK: if we didn't skip ahead, arg didn't match:
		HelpInvalidArg(sub, arg)
		Help(sub, summary, opts)
	}
}
