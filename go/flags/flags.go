package flags

import (
	"fmt"
	"reflect"
	"strings"
)

const PARSE_ERROR_PREFIX = "internal err: ParseArgs"

type ParseError struct {
	message string
}

func (pe ParseError) Error() string {
	return fmt.Sprintf("%s: %s", PARSE_ERROR_PREFIX, pe.message)
}

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
func expandOpts(opts any) ([]Flag, error) {
	flags := []Flag{}
	v := reflect.ValueOf(opts)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	} else {
		return nil, ParseError{fmt.Sprintf("needs a pointer to a struct: %#v", opts)}
	}
	if v.Kind() != reflect.Struct {
		return nil, ParseError{fmt.Sprintf("needs an underlying struct: %#v", opts)}
	}
	for j := 0; j < v.NumField(); j++ {
		field := v.Field(j)
		if field.Kind() != reflect.Struct {
			return nil, ParseError{fmt.Sprintf("needs a struct: %#v", field.Kind())}
		}
		if !field.CanAddr() {
			return nil, ParseError{fmt.Sprintf("needs an addressable type: %#v", field.Type())}
		}
		// NOTE: Addr() gets the underlying address:
		if f, ok := field.Addr().Interface().(Flag); ok {
			flags = append(flags, f)
		} else {
			return nil, ParseError{fmt.Sprintf("needs a []Flag interface: %#v", field.Type())}
		}
	}
	return flags, nil
}

// internal for type safety testing
func parseArgsInternal(help HelpError, args []string, opts any, flags []Flag) error {
argloop:
	for i := 0; i < len(args); i++ {
		arg := args[i]
		for j, f := range flags {
			if f.Match(arg, i, j) {
				if f.IsHelp() {
					// so Help.Val is true to avoid confusion:
					f.Set(arg)
					return help.Help(opts)
				}
				if !f.Valid(args, i) {
					return help.HelpMissingAssignment(arg)
				}
				i = f.MaybeIncrement(i)
				f.Set(args[i])
				continue argloop
			}
		}
		// NOTE: if we didn't skip ahead, arg didn't match:
		return help.HelpInvalidArg(arg)
	}
	return nil
}

// assign values of args to opts struct using Flag interface methods
func ParseArgs(help HelpError, args []string, opts any) error {
	flags, err := expandOpts(opts)
	if err != nil {
		return err
	}
	return parseArgsInternal(help, args, opts, flags)
}
