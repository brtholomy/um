package tag

import (
	"log"

	"github.com/brtholomy/um/go/flags"
)

const cmd = "tag"

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
		flags.Flag[string]{"--date", "-d", "", "date range"},
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
	log.Println(opts)
}
