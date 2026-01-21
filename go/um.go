package main

import (
	"fmt"
	"log"
	"os"

	"github.com/brtholomy/um/go/cat"
	"github.com/brtholomy/um/go/cmd"
	"github.com/brtholomy/um/go/last"
	"github.com/brtholomy/um/go/next"
	"github.com/brtholomy/um/go/sort"
	"github.com/brtholomy/um/go/tag"
)

var helpShort string = fmt.Sprintf("um [%s | %s | %s | %s | %s]", cmd.Tag, cmd.Next, cmd.Last, cmd.Sort, cmd.Help)
var helpLong string = fmt.Sprintf(`%s

An (U)ltralight database for (M)arkdown composition.

Each subcommand has a --help | -h flag.

https://github.com/brtholomy/um
`, helpShort)

func main() {
	// NOTE: no prefix at all so that I can use log.Fatalf
	log.SetFlags(0)

	// TODO: do something useful:
	// run empty tag query?
	if len(os.Args) < 2 {
		log.Fatalln(helpShort)
	}

	arg := cmd.Subcommand(os.Args[1])
	args := os.Args[2:]

	switch arg {
	case cmd.Next:
		next.Next(args)
	case cmd.Last:
		last.Last(args)
	case cmd.Tag:
		tag.Tag(args)
	case cmd.Sort:
		sort.Sort(args)
	case cmd.Cat:
		cat.Cat(args)
	case cmd.Help:
		log.Fatalln(helpLong)
	default:
		log.Println("um: command not found")
		log.Fatalln(helpShort)
	}
}
