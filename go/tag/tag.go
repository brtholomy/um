package tag

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/brtholomy/um/go/cmd"
	"github.com/brtholomy/um/go/flags"
	"github.com/brtholomy/um/go/last"
)

const (
	CMD     = cmd.Tag
	SUMMARY = "query for a filelist with a set of tags"
)

type options struct {
	Query   flags.Arg
	Date    flags.String
	Invert  flags.Bool
	Verbose flags.Bool
	Help    flags.Bool
}

func initOpts() options {
	return options{
		flags.Arg{"", "tag query: understands intersection '+' and union ','"},
		flags.String{"--date", "-d", "", "date range in ISO 8601: YYYY.MM.DD[-YYYY.MM.DD]"},
		flags.Bool{"--invert", "-i", false, "invert match"},
		flags.Bool{"--verbose", "-v", false, "print a verbose summary"},
		flags.Bool{"--help", "-h", false, "show help"},
	}
}

func Tag(args []string) {
	opts := initOpts()
	if err := flags.ParseArgs(CMD, SUMMARY, args, &opts); err != nil {
		var herr flags.HelpError
		if errors.As(err, &herr) {
			fmt.Println(herr)
			return
		}
		log.Fatalf("um %s: %s", CMD, err)
	}

	queries := parseQuery(opts.Query.Val)
	entries := entriesGlobOrStdin(last.GLOB)

	// we shrink the entries list immediately if we want a date range:
	if opts.Date.IsSet() {
		entries = dateRange(entries, opts.Date.Val)
	}
	tagmap := makeTagmap(entries)

	// processQueries must precede invert because we want invert to respect combined tags:
	files := processQueries(tagmap, queries)
	if opts.Invert.IsSet() {
		files = invert(entries, files)
	}
	// NOTE: the full makeAdjacencies map may one day be useful on its own
	adjacencies := reduceAdjacencies(makeAdjacencies(entries, files), queries, opts.Invert.Val)

	printFiles(os.Stdout, entries, tagmap, files, adjacencies, queries, opts.Verbose.Val)
}
