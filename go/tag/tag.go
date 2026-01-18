package tag

import (
	"fmt"
	"reflect"
	"strings"
)

var opts struct {
	query  string `name:"query" help:"tag query"`
	invert bool   `name:"invert" help:"invert match"`
}

func stripDashes(flag string) string {
	return strings.TrimPrefix(flag, "--")
}

func Tag(args []string) {

	// reflection!
	t := reflect.TypeOf(opts)
	field := reflect.StructField{}
	found := false
	for _, a := range args {
		field, found = t.FieldByName(stripDashes(a))
		if !found {
			fmt.Printf("%s not found in opts\n", a)
			return
		}
	}
	tag := field.Tag

	fmt.Printf("field name: %#v\n", tag.Get("name"))
	fmt.Printf("field help: %#v\n", tag.Get("help"))

	// for i := 0; i < t.NumField(); i++ {
	// 	field := t.Field(i)
	// 	if alias, ok := field.Tag.Lookup("name"); ok {
	// 		if alias == "" {
	// 			fmt.Println("(blank)")
	// 		} else {
	// 			fmt.Println(alias)
	// 		}
	// 	} else {
	// 		fmt.Println("(not specified)")
	// 	}
	// }
}
