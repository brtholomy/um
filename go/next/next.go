package next

import (
	"fmt"
	"os"

	"github.com/brtholomy/um/go/flags"
)

type options struct {
	descriptor flags.Arg
	tags       flags.Arg
}

func InitOpts() options {
	return options{
		flags.Arg{"", "midfix file descriptor"},
		flags.Arg{"", "tags to add to new file"},
	}
}

func ParseArgs(args []string) options {
	opts := InitOpts()
	for i, arg := range args {
		switch {
		case i == 0 && !flags.HasDashPrefix(arg):
			opts.descriptor.Val = arg
		case i == 1 && !flags.HasDashPrefix(arg):
			opts.tags.Val = arg
		case i >= 2:
			fmt.Println("um next : too many args")
			os.Exit(1)
		}
	}
	return opts
}

func Next(args []string) {
	fmt.Println("next")
	opts := ParseArgs(args)
	fmt.Println(opts)
}
