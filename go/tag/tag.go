package tag

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"
)

var opts struct {
	Query  string `name:"query" help:"tag query"`
	Invert bool   `name:"invert" help:"invert match"`
}

func flagToField(flag string) string {
	s := strings.TrimPrefix(flag, "--")
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

func Tag(args []string) {

	// reflection!
	t := reflect.TypeOf(opts)
	field := reflect.StructField{}
	found := false
	for _, a := range args {
		field, found = t.FieldByName(flagToField(a))
		if !found {
			fmt.Printf("%s not found in opts\n", a)
			return
		}
	}
	tag := field.Tag

	fmt.Printf("field name: %#v\n", tag.Get("name"))
	fmt.Printf("field help: %#v\n", tag.Get("help"))

	v := reflect.ValueOf(opts)
	vfield := reflect.Value{}
	for _, a := range args {
		vfield = v.FieldByName(flagToField(a))
		if vfield.IsValid() && vfield.CanSet() && vfield.Kind() == reflect.Bool {
			vfield.SetBool(true)
			fmt.Printf("vfield: %#v\n", vfield)
		}
	}
	fmt.Printf("vfield: %#v\n", vfield)
	fmt.Printf("opts: %#v\n", opts)
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
