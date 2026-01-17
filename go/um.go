package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

type Subcommand string

const (
	tag  Subcommand = "tag"
	next Subcommand = "next"
	last Subcommand = "last"
	sort Subcommand = "sort"
)

var help string = fmt.Sprintf("um [%s | %s | %s | %s]", tag, next, last, sort)

type Flag string

const (
	query  Flag = "query"
	invert Flag = "invert"
	key    Flag = "key"
)

func (f Flag) Dash() string {
	return fmt.Sprintf("--%s", f)
}

// prepend --flag to args if a flag isn't there:
// allows trailing flags:
// um tag foo --invert
// becomes:
// um tag --query foo --invert
func prependFlagToArgs(args []string, flag Flag) []string {
	if len(args) >= 1 && !strings.HasPrefix(args[0], "-") {
		args = append([]string{flag.Dash()}, args...)
	}
	return args
}

func main() {
	tagFlags := flag.NewFlagSet(string(tag), flag.ExitOnError)
	tagQuery := tagFlags.String(string(query), "", "tag query")
	tagInvert := tagFlags.Bool(string(invert), false, "invert match")

	sortFlags := flag.NewFlagSet(string(sort), flag.ExitOnError)
	sortKey := sortFlags.String(string(key), "", "filename of ordered list to use as key")

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
		fmt.Printf("args: %#v\n", args)
	case tag:
		tagFlags.Parse(prependFlagToArgs(args, query))
		fmt.Println(tag)
		fmt.Printf("query: %#v\n", *tagQuery)
		fmt.Printf("invert: %#v\n", *tagInvert)
		fmt.Printf("args: %#v\n", tagFlags.Args())
		tagFlags.Usage()
	case sort:
		sortFlags.Parse(args)
		fmt.Println("subcmd sort")
		fmt.Printf("key: %#v\n", *sortKey)
		fmt.Printf("args: %#v\n", sortFlags.Args())
		sortFlags.Usage()
	default:
		fmt.Println(help)
		os.Exit(1)
	}
}
