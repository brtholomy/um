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
	message string
}

func (h HelpError) Error() string {
	return h.message
}

func HelpRequired(sub cmd.Subcommand, long string) error {
	return HelpError{
		fmt.Sprintf("um %s: %s is required", sub, long),
	}
}

func HelpInvalidArg(sub cmd.Subcommand, arg string) error {
	return HelpError{
		fmt.Sprintf("um %s: invalid argument: %s", sub, arg),
	}
}

func HelpMissingAssignment(sub cmd.Subcommand, arg string) error {
	return HelpError{
		fmt.Sprintf("um %s: %s needs a value assignment", sub, arg),
	}
}

// print out help string by reflecting over fields of provided opts struct
// and os.Exit(0)
func Help(sub cmd.Subcommand, summary string, opts any) error {
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
	if _, err := buf.WriteString(fmt.Sprintf("um %s%s\n\n%s\n\n", sub, positional, summary)); err != nil {
		return err
	}
	w.Flush()
	return HelpError{buf.String()}
}
