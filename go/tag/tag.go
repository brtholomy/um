package tag

import (
	"flag"
	"fmt"

	"github.com/brtholomy/um/go/flags"
)

const (
	query  flags.Flag = "query"
	invert flags.Flag = "invert"
)

func Tag(args []string) {
	tagFlags := flag.NewFlagSet("tag", flag.ExitOnError)
	tagQuery := tagFlags.String(string(query), "", "tag query")
	tagInvert := tagFlags.Bool(string(invert), false, "invert match")

	tagFlags.Parse(flags.PrependFlagToArgs(args, query))

	fmt.Printf("query: %#v\n", *tagQuery)
	fmt.Printf("invert: %#v\n", *tagInvert)
	fmt.Printf("args: %#v\n", tagFlags.Args())
	tagFlags.Usage()
}
