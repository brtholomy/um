package sort

import (
	"fmt"

	"github.com/brtholomy/um/go/flags"
)

type options struct {
	Source flags.Flag[string]
	Key    flags.Flag[string]
	Help   flags.Flag[bool]
}

func initOpts() options {
	return options{
		flags.Flag[string]{"--source", "-s", "", "path to source list. reads from stdin if omitted."},
		flags.Flag[string]{"--key", "-k", "", "path to sort key"},
		flags.Flag[bool]{"--help", "-h", false, "show help"},
	}
}

func parseArgs(args []string) options {
	opts := initOpts()
	for i, arg := range args {
		switch {
		case arg == opts.Source.Long || arg == opts.Source.Short:
			if flags.MissingValueArg(args, i) {
				flags.HelpMissingVal(arg)
				break
			}
			opts.Source.Val = args[i+1]
		case arg == opts.Key.Long || arg == opts.Key.Short:
			if flags.MissingValueArg(args, i) {
				flags.HelpMissingVal(arg)
				break
			}
			opts.Key.Val = args[i+1]
		case arg == opts.Help.Long || arg == opts.Help.Short:
			flags.Help("sort", opts)
			return opts
		}
	}
	return opts
}

func Sort(args []string) {
	opts := parseArgs(args)
	fmt.Println(opts)
}
