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

func HelpRequired(sub cmd.Subcommand, long string) {
	log.Printf("um %s: %s is required", sub, long)
}

func HelpInvalidArg(sub cmd.Subcommand, arg string) {
	log.Printf("um %s: invalid argument: %s", sub, arg)
}

func HelpMissingAssignment(sub cmd.Subcommand, arg string) {
	log.Printf("um %s: %s needs a value assignment\n", sub, arg)
	log.Fatal("try: um [cmd] --help")
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
	log.Printf("um %s%s\n\n", sub, positional)
	log.Printf("%s\n\n", summary)
	w.Flush()
	os.Exit(0)
}
