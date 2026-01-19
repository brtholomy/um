package next

import (
	"fmt"

	"github.com/brtholomy/um/go/flags"
)

type options struct {
	Descriptor flags.Arg
	Tags       flags.Arg
	Help       flags.Flag[bool]
}

func initOpts() options {
	return options{
		flags.Arg{"", "midfix file descriptor"},
		flags.Arg{"", "tags to add to new file"},
		flags.Flag[bool]{"--help", "-h", false, "show help"},
	}
}

func parseArgs(args []string) options {
	opts := initOpts()
	for i, arg := range args {
		switch {
		case i == 0 && !flags.HasDashPrefix(arg):
			opts.Descriptor.Val = arg
		case i == 1 && !flags.HasDashPrefix(arg):
			opts.Tags.Val = arg
		case i >= 2:
			fmt.Println("um next : too many args")
			flags.Help("next", opts)
		case arg == opts.Help.Long || arg == opts.Help.Short:
			flags.Help("next", opts)
		}
	}
	return opts
}

func Next(args []string) {
	opts := parseArgs(args)
	fmt.Println(opts)
}
