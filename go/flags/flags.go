package flags

import (
	"log"
	"reflect"
	"strings"

	"github.com/brtholomy/um/go/cmd"
)

const INTERNAL_ERR_PREFIX = "um internal err: ParseArgs"

// interface as func parameter
type Flag interface {
	SetValid(args []string, i int) (int, bool)
	Match(arg string, i, j int) bool
	IsSet() bool
	IsHelp() bool
}

type Arg struct {
	Val  string
	Help string
}

type String struct {
	Long  string
	Short string
	Val   string
	Help  string
}

type Bool struct {
	Long  string
	Short string
	Val   bool
	Help  string
}

func (f *Arg) SetValid(args []string, i int) (int, bool) {
	f.Val = args[i]
	return i, true
}

func (f *Arg) Match(arg string, i, j int) bool {
	return i == j && !hasDashPrefix(arg)
}

func (f *Arg) IsSet() bool {
	return f.Val != ""
}

func (f *Arg) IsHelp() bool {
	return false
}

func (f *String) SetValid(args []string, i int) (int, bool) {
	i, val, ok := validateIncrementFetch(args, i)
	f.Val = val
	return i, ok
}

func (f *String) Match(arg string, _, _ int) bool {
	return arg == f.Long || arg == f.Short
}

func (f *String) IsSet() bool {
	return f.Val != ""
}

func (f *String) IsHelp() bool {
	return false
}

func (f *Bool) SetValid(_ []string, i int) (int, bool) {
	f.Val = true
	return i, true
}

func (f *Bool) Match(arg string, _, _ int) bool {
	return arg == f.Long || arg == f.Short
}

func (f *Bool) IsSet() bool {
	return f.Val
}

func (f *Bool) IsHelp() bool {
	return f.Long == "--help"
}

func hasDashPrefix(s string) bool {
	return strings.HasPrefix(s, "-")
}

// check for non-dashed value ahead in the args slice
func missingValue(args []string, i int) bool {
	return i+1 == len(args) || hasDashPrefix(args[i+1])
}

// 1. validate a Flag[string] by looking ahead in args, if missing, return false
// 2. increment the i to skip that value and return
// 3. fetch that value and return
func validateIncrementFetch(args []string, i int) (int, string, bool) {
	if missingValue(args, i) {
		return 0, "", false
	}
	return i + 1, args[i+1], true
}

// expand incoming opts struct into a []Flag
//
// WARN: incoming opts must be a pointer!
// WORB: opts struct must list its Arg types first and in expected order
// NOTE: uglier than the pretty switch statements I had per cmd, but much more convenient.
func expandOpts(opts any) []Flag {
	flags := []Flag{}
	v := reflect.ValueOf(opts)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	} else {
		log.Fatalf("%s needs a pointer to a struct: %#v", INTERNAL_ERR_PREFIX, opts)
	}
	if v.Kind() != reflect.Struct {
		log.Fatalf("%s needs an underlying struct: %#v", INTERNAL_ERR_PREFIX, opts)
	}
	for j := 0; j < v.NumField(); j++ {
		field := v.Field(j)
		if field.Kind() != reflect.Struct {
			log.Fatalf("%s needs a struct: %s", INTERNAL_ERR_PREFIX, field.Kind())
		}
		if !field.CanAddr() {
			log.Fatalf("%s needs an addressable type: ", INTERNAL_ERR_PREFIX, field.Type())
		}
		// NOTE: Addr() gets the underlying address:
		if f, ok := field.Addr().Interface().(Flag); ok {
			flags = append(flags, f)
		}
	}
	return flags
}

// assign values of args to opts struct using Flag interface methods
func ParseArgs(sub cmd.Subcommand, summary string, args []string, opts any) {
	flags := expandOpts(opts)
argloop:
	for i := 0; i < len(args); i++ {
		arg := args[i]
		for j, f := range flags {
			if f.Match(arg, i, j) {
				if f.IsHelp() {
					Help(sub, summary, opts)
				}
				ok := false
				i, ok = f.SetValid(args, i)
				if !ok {
					log.Printf("um %s: %s needs a value assignment\n", sub, args[i])
					log.Fatal("try: um [cmd] --help")
				}
				continue argloop
			}
		}
		// NOTE: if we didn't skip ahead, arg didn't match:
		HelpInvalidArg(sub, arg)
		Help(sub, summary, opts)
	}
}
