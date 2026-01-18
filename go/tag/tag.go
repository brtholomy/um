package tag

import (
	"fmt"

	"github.com/brtholomy/um/go/flags"
)

type options struct {
	Query  flags.Arg
	Date   flags.Flag[string]
	Invert flags.Flag[bool]
	Help   flags.Flag[bool]
}

func InitOpts() options {
	return options{
		flags.Arg{"", "tag query"},
		flags.Flag[string]{"--date", "-d", "", "date range"},
		flags.Flag[bool]{"--invert", "-i", false, "invert match"},
		flags.Flag[bool]{"--help", "-h", false, "show help"},
	}
}

func ParseArgs(args []string) options {
	opts := InitOpts()
	for i, arg := range args {
		switch {
		case i == 0 && !flags.HasDashPrefix(arg):
			opts.Query.Val = arg
		case arg == opts.Invert.Long || arg == opts.Invert.Short:
			opts.Invert.Val = true
		case arg == opts.Date.Long || arg == opts.Date.Short:
			if flags.MissingValueArg(args, i) {
				flags.HelpMissingVal(arg)
				break
			}
			opts.Date.Val = args[i+1]
		case arg == opts.Help.Long || arg == opts.Help.Short:
			flags.Help("tag", opts)
			return opts
		}
	}
	return opts
}

func Tag(args []string) {
	opts := ParseArgs(args)
	fmt.Println(opts)
}
