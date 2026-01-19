package main

import (
	"fmt"
	"log"
	"os"

	nextpkg "github.com/brtholomy/um/go/next"
	sortpkg "github.com/brtholomy/um/go/sort"
	tagpkg "github.com/brtholomy/um/go/tag"
)

type Subcommand string

const (
	tag  Subcommand = "tag"
	next Subcommand = "next"
	last Subcommand = "last"
	sort Subcommand = "sort"
	help Subcommand = "help"
)

var helpShort string = fmt.Sprintf("um [%s | %s | %s | %s | %s]", tag, next, last, sort, help)
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

	cmd := Subcommand(os.Args[1])
	args := os.Args[2:]

	switch cmd {
	case next:
		nextpkg.Next(args)
	case last:
		nextpkg.Last()
	case tag:
		tagpkg.Tag(args)
	case sort:
		sortpkg.Sort(args)
	case help:
		log.Fatalln(helpLong)
	default:
		log.Println("um: command not found")
		log.Fatalln(helpShort)
	}
}
