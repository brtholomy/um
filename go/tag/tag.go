package tag

import (
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
	Date    flags.Flag[string]
	Invert  flags.Flag[bool]
	Verbose flags.Flag[bool]
	Help    flags.Flag[bool]
}

func initOpts() options {
	return options{
		flags.Arg{"", "tag query: understands intersection '+' and union ','"},
		flags.Flag[string]{"--date", "-d", "", "date range in ISO 8601: YYYY.MM.DD[-YYYY.MM.DD]"},
		flags.Flag[bool]{"--invert", "-i", false, "invert match"},
		flags.Flag[bool]{"--verbose", "-v", false, "print a verbose summary"},
		flags.Flag[bool]{"--help", "-h", false, "show help"},
	}
}

func Tag(args []string) {
	opts := initOpts()
	flags.ParseArgs(CMD, SUMMARY, args, &opts)

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
