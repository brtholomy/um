package sort

import (
	"log"

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
			flags.ValidValueOrExit(args, i)
			opts.Source.Val = args[i+1]
		case arg == opts.Key.Long || arg == opts.Key.Short:
			flags.ValidValueOrExit(args, i)
			opts.Key.Val = args[i+1]
		case arg == opts.Help.Long || arg == opts.Help.Short:
			flags.Help("sort", opts)
		default:
			log.Printf("um sort: invalid argument: %s", arg)
			flags.Help("sort", opts)
		}
	}
	return opts
}

func Sort(args []string) {
	opts := parseArgs(args)
	log.Println(opts)
}
