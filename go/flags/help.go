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
	t := v.Type()

	buf := &bytes.Buffer{}
	w := tabwriter.NewWriter(buf, 0, 0, 3, ' ', 0)
	positional := ""

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)
		switch value.Interface().(type) {
		case Arg:
			positional += fmt.Sprintf(" [%s]", strings.ToLower(field.Name))
			fmt.Fprintf(w, "[%v]\tstring\t%v\n", strings.ToLower(field.Name), value.FieldByName("Help"))
		case String, Bool:
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
	if _, err := buf.WriteString(fmt.Sprintf("um %s%s\n\n%s\n\n", h.sub, positional, h.summary)); err != nil {
		return err
	}
	w.Flush()
	h.message = buf.String()
	return h
}
