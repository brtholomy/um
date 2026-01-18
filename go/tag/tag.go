package tag

import (
	"fmt"

	"github.com/brtholomy/um/go/flags"
)

type options struct {
	query  flags.Arg
	date   flags.Flag
	invert flags.FlagBool
}

func InitOpts() options {
	return options{
		flags.Arg{"", "tag query"},
		flags.Flag{"--date", "-d", "", "date range"},
		flags.FlagBool{"--invert", "-i", false, "invert match"},
	}
}

func ParseArgs(args []string) options {
	opts := InitOpts()
	if len(args) == 0 {
		return opts
	}
	if flags.HasDashPrefix(args[0]) {
		opts.query.Val = args[0]
	}
	for i, arg := range args {
		switch arg {
		case opts.invert.Long, opts.invert.Short:
			opts.invert.Val = true
		case opts.date.Long:
			if flags.MissingValueArg(args, i) {
				flags.HelpMissingVal(arg)
				break
			}
			opts.date.Val = args[i+1]
		}
	}
	return opts
}

func Tag(args []string) {
	fmt.Printf("args: %#v\n", args)
	opts := ParseArgs(args)
	fmt.Printf("opts: %#v\n", opts)
}
