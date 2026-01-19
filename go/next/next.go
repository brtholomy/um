package next

import (
	"errors"
	"fmt"
	"strconv"

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

func parseFile(last string) (string, string, error) {
	res := fileRegexp.FindStringSubmatch(last)
	num := ""
	desc := ""
	if len(res) < 2 {
		return num, desc, errors.New("no um files in current dir")
	}
	num = res[1]
	if len(res) > 2 {
		desc = res[2]
	}
	return num, desc, nil
}

func next(last string) (string, error) {
	num, desc, err := parseFile(last)
	if err != nil {
		return "", err
	}
	i, err := strconv.Atoi(num)
	if err != nil {
		return "", err
	}
	if desc != "" {
		desc = desc + "."
	}
	width := len(num)
	fmtstr := fmt.Sprintf("%%0%dd.%%smd", width)
	n := fmt.Sprintf(fmtstr, i+1, desc)
	return n, nil
}

func Next(args []string) {
	opts := parseArgs(args)
	fmt.Println(opts)

	l, err := last()
	if err != nil {
		fmt.Printf("err: %v\n", err)
	}
	n, err := next(l)
	if err != nil {
		fmt.Printf("err: %v\n", err)
	}
	fmt.Println(n)
}
