package flags

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"text/tabwriter"

	"github.com/brtholomy/um/go/cmd"
)

type HelpError struct {
	sub     cmd.Subcommand
	summary string
	message string
}

func (h HelpError) Error() string {
	return h.message
}

// TODO: is this odd? Constructing an error at the callsite and passing it down and back up?
// but how else should I fill out its fields to be reused later?
func NewHelpError(sub cmd.Subcommand, summary string) HelpError {
	return HelpError{sub, summary, ""}
}

func (h HelpError) HelpRequired(long string) error {
	h.message = fmt.Sprintf("um %s: %s is required", h.sub, long)
	return h
}

func (h HelpError) HelpInvalidArg(arg string) error {
	h.message = fmt.Sprintf("um %s: invalid argument: %s", h.sub, arg)
	return h

}

func (h HelpError) HelpMissingAssignment(arg string) error {
	h.message = fmt.Sprintf("um %s: %s needs a value assignment", h.sub, arg)
	return h
}

// assemble --help string by reflecting over fields of provided opts struct
func (h HelpError) Help(opts any) error {
	v := reflect.ValueOf(opts)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return ParseError{fmt.Sprintf("needs an underlying struct: %#v", opts)}
	}
	t := v.Type()

	buf := &bytes.Buffer{}
	w := tabwriter.NewWriter(buf, 0, 0, 3, ' ', 0)
	positional := ""

	for i := 0; i < v.NumField(); i++ {
		tField := t.Field(i)
		vField := v.Field(i)
		switch vField.Interface().(type) {
		case Arg:
			positional += fmt.Sprintf(" [%s]", strings.ToLower(tField.Name))
			fmt.Fprintf(w, "[%v]\tstring\t%v\n", strings.ToLower(tField.Name), vField.FieldByName("Help"))
		case String, Bool:
			if tField.Name != "Help" {
				positional += fmt.Sprintf(" [%s]", vField.FieldByName("Long"))
			}
			fmt.Fprintf(w, "%v | %v\t%v\t%v\n",
				vField.FieldByName("Long"),
				vField.FieldByName("Short"),
				vField.FieldByName("Val").Type(),
				vField.FieldByName("Help"),
			)
		default:
			return ParseError{fmt.Sprintf("needs a []Flag interface: %#v", vField.Type())}

		}
	}
	if _, err := buf.WriteString(fmt.Sprintf("um %s%s\n\n%s\n\n", h.sub, positional, h.summary)); err != nil {
		return err
	}
	if err := w.Flush(); err != nil {
		return err
	}
	h.message = buf.String()
	return h
}
