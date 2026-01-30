package main

import (
	"fmt"
	"log"
	"os"

	"github.com/brtholomy/um/go/cat"
	"github.com/brtholomy/um/go/cmd"
	"github.com/brtholomy/um/go/last"
	"github.com/brtholomy/um/go/mv"
	"github.com/brtholomy/um/go/next"
	"github.com/brtholomy/um/go/sort"
	"github.com/brtholomy/um/go/tag"
)

var helpShort string = fmt.Sprintf("um [%s | %s | %s | %s | %s | %s | %s]", cmd.Next, cmd.Last, cmd.Tag, cmd.Cat, cmd.Sort, cmd.Mv, cmd.Help)
var helpLong string = fmt.Sprintf(`%s

(U)ltralight zettelkasten for (M)arkdown composition.

Each subcommand has a --help | -h flag.

https://github.com/brtholomy/um
`, helpShort)

func main() {
	// NOTE: no prefix at all so that I can use log alongside fmt
	// log : stderr : exit status 1
	// fmt : stdout : exit status 0
	log.SetFlags(0)

	if len(os.Args) < 2 {
		log.Fatalln(helpShort)
	}

	arg := cmd.Subcommand(os.Args[1])
	// NOTE: just pass what's relevant:
	args := os.Args[2:]

	switch arg {
	case cmd.Next:
		next.Next(args)
	case cmd.Last:
		last.Last(args)
	case cmd.Tag:
		tag.Tag(args)
	case cmd.Cat:
		cat.Cat(args)
	case cmd.Sort:
		sort.Sort(args)
	case cmd.Mv:
		mv.Mv(args)
	case cmd.Help:
		fmt.Println(helpLong)
	default:
		log.Println("um: command not found")
		log.Fatalln(helpShort)
	}
}
