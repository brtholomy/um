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
	Set(string)
	Valid([]string, int) bool
	MaybeIncrement(int) int
	Match(string, int, int) bool
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

func (f *Arg) Set(arg string) {
	f.Val = arg
}

func (f *Arg) Valid(args []string, i int) bool {
	return true
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

func (f *Arg) MaybeIncrement(i int) int {
	return i
}

func (f *String) Set(arg string) {
	f.Val = arg
}

func (f *String) Valid(args []string, i int) bool {
	return !missingValue(args, i)
}

func (f *String) MaybeIncrement(i int) int {
	return i + 1
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

func (f *Bool) Set(arg string) {
	f.Val = true
}

func (f *Bool) Match(arg string, _, _ int) bool {
	return arg == f.Long || arg == f.Short
}

func (f *Bool) Valid(args []string, i int) bool {
	return true
}

func (f *Bool) MaybeIncrement(i int) int {
	return i
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
			log.Fatalf("%s needs an addressable type: %s", INTERNAL_ERR_PREFIX, field.Type())
		}
		// NOTE: Addr() gets the underlying address:
		if f, ok := field.Addr().Interface().(Flag); ok {
			flags = append(flags, f)
		} else {
			log.Fatalf("%s needs a []Flag interface: %s", INTERNAL_ERR_PREFIX, field.Type())
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
				if !f.Valid(args, i) {
					HelpMissingAssignment(sub, arg)
				}
				i = f.MaybeIncrement(i)
				f.Set(args[i])
				continue argloop
			}
		}
		// NOTE: if we didn't skip ahead, arg didn't match:
		HelpInvalidArg(sub, arg)
		Help(sub, summary, opts)
	}
}
