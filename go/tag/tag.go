package tag

import (
	"os"

	cmdpkg "github.com/brtholomy/um/go/cmd"
	"github.com/brtholomy/um/go/flags"
)

// just so that copied calls into flags.Help don't have to be adjusted between files:
const cmd = cmdpkg.Tag

type options struct {
	Query   flags.Arg
	Date    flags.Flag[string]
	Invert  flags.Flag[bool]
	Verbose flags.Flag[bool]
	Help    flags.Flag[bool]
}

func initOpts() options {
	return options{
		flags.Arg{"", "tag query"},
		flags.Flag[string]{"--date", "-d", "", "date range in ISO 8601: YYYY.MM.DD[-YYYY.MM.DD]"},
		flags.Flag[bool]{"--invert", "-i", false, "invert match"},
		flags.Flag[bool]{"--verbose", "-v", false, "print a verbose summary"},
		flags.Flag[bool]{"--help", "-h", false, "show help"},
	}
}

func parseArgs(args []string) options {
	opts := initOpts()
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case i == 0 && !flags.HasDashPrefix(arg):
			opts.Query.Val = arg
		case arg == opts.Invert.Long || arg == opts.Invert.Short:
			opts.Invert.Val = true
		case arg == opts.Verbose.Long || arg == opts.Verbose.Short:
			opts.Verbose.Val = true
		case arg == opts.Date.Long || arg == opts.Date.Short:
			i, opts.Date.Val = flags.ValidateIncrementFetchOrExit(args, i)
		case arg == opts.Help.Long || arg == opts.Help.Short:
			flags.Help(cmd, opts)
		default:
			flags.HelpInvalidArg(cmd, arg)
			flags.Help(cmd, opts)
		}
	}
	return opts
}

func Tag(args []string) {
	opts := parseArgs(args)
	queries := parseQuery(opts.Query.Val)
	entries := entriesGlobOrStdin()

	// we shrink the entries list immediately if we want a date range:
	if opts.Date.Val != "" {
		entries = dateRange(entries, opts.Date.Val)
	}
	tagmap := makeTagmap(entries)

	// processQueries must precede invert because we want invert to respect combined tags:
	files := processQueries(tagmap, queries)
	if opts.Invert.Val {
		files = invert(entries, files)
	}
	// NOTE: the full makeAdjacencies map may one day be useful on its own
	adjacencies := reduceAdjacencies(makeAdjacencies(entries, files), queries, opts.Invert.Val)

	printFiles(os.Stdout, entries, tagmap, files, adjacencies, queries, opts.Verbose.Val)
}
