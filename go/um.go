package main

import (
	"fmt"
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
)

var help string = fmt.Sprintf("um [%s | %s | %s | %s] [--help | -h]", tag, next, last, sort)

func main() {
	// TODO: do something useful:
	// run empty tag query?
	if len(os.Args) < 2 {
		fmt.Println(help)
		os.Exit(1)
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
	default:
		fmt.Println("command not found")
		fmt.Println(help)
		os.Exit(1)
	}
}
