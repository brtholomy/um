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
	Match(arg string, i, j int) bool
	IsSet() bool
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

func (f *Arg) Set(val string) {
	f.Val = val
}

func (f *Arg) Match(arg string, i, j int) bool {
	return i == j && !hasDashPrefix(arg)
}

func (f *Arg) IsSet() bool {
	return f.Val != ""
}

func (f *String) Set(val string) {
	f.Val = val
}

// TODO: this should look ahead
// i, fetch = ValidateIncrementFetchOrExit(sub, args, i)
func (f *String) Match(arg string, _, _ int) bool {
	return arg == f.Long || arg == f.Short
}

func (f *String) IsSet() bool {
	return f.Val != ""
}

func (f *Bool) Set(_ string) {
	f.Val = true
}

func (f *Bool) Match(arg string, _, _ int) bool {
	// TODO: catch --help
	return arg == f.Long || arg == f.Short
}

func (f *Bool) IsSet() bool {
	return f.Val
}

func hasDashPrefix(s string) bool {
	return strings.HasPrefix(s, "-")
}

// check for non-dashed value ahead in the args slice
func missingValue(args []string, i int) bool {
	return i+1 == len(args) || hasDashPrefix(args[i+1])
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

// expand incoming opts struct into a []Flag
//
// WARN: incoming opts must be a pointer!
// WORB: opts struct must list its Arg types first and in expected order
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
//
// NOTE: uglier than the pretty switch statements I had per cmd, but much more convenient.
func ParseArgs(sub cmd.Subcommand, summary string, args []string, opts any) {
	flags := expandOpts(opts)
argloop:
	for i, arg := range args {
		for j, f := range flags {
			if f.Match(arg, i, j) {
				f.Set(arg)
				continue argloop
			}
		}
		// NOTE: if we didn't skip ahead, arg didn't match:
		HelpInvalidArg(sub, arg)
		Help(sub, summary, opts)
	}
}
