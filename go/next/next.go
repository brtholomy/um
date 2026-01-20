package next

import (
	"errors"
	"fmt"
	"log"
	"os/exec"
	"strconv"

	cmdpkg "github.com/brtholomy/um/go/cmd"
	"github.com/brtholomy/um/go/flags"
)

const cmd = cmdpkg.Next

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
		case arg == opts.Help.Long || arg == opts.Help.Short:
			flags.Help(cmd, opts)
		default:
			flags.HelpInvalidArg(cmd, arg)
			flags.Help(cmd, opts)
		}
	}
	return opts
}

// takes the complete last file string
// returns the number as string
func numFromLast(last string) (string, error) {
	res := fileRegexp.FindStringSubmatch(last)
	num := ""
	if len(res) < 2 {
		return num, errors.New(NOT_FOUND_MSG)
	}
	num = res[1]
	return num, nil
}

// takes the complete last file string and new descriptor
// returns the complete next file string
func next(last string, desc string) (string, error) {
	num, err := numFromLast(last)
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
	// whaddya doin bro wahh
	fmtstr := fmt.Sprintf("%%0%dd.%%smd", width)
	n := fmt.Sprintf(fmtstr, i+1, desc)
	return n, nil
}

// calls (um-next) inside a running emacs server. which has the advantage of fetching the loaded
// value of um-date-format, and respecting any other loaded configuration we might add.
//
// other options would be to forego respecting Elisp config values, or loading from a common config
// file and forcing emacs to load from there - which would be very un-Emacs. But as the file loading
// assumes a running server anyway, might as well get the current values.
//
// this is the point at which this design becomes weird. But I don't want to let Elisp suck this CLI
// into its nasty grip. Go is far superior in this application and there's no way I'm writing the
// `um tag` logic in Elisp.
//
// So I've got Go and and an ancient beloved Lisp machine trying to live together.
func emacsNext(f string, tags string) error {
	quotedargs := fmt.Sprintf(`"%s"`, f)
	if tags != "" {
		quotedargs = fmt.Sprintf(`%s "%s"`, quotedargs, tags)
	}
	fn := fmt.Sprintf(`(um-next %s)`, quotedargs)
	out, err := exec.Command("emacsclient", "--eval", fn).CombinedOutput()
	if err != nil {
		return err
	}
	// NOTE: sends to stderr : is this what we want?
	// if we decide to pipe the filename out, yes.
	log.Print(string(out))
	return nil
}

func Next(args []string) {
	opts := parseArgs(args)
	l, err := last()
	if err != nil {
		log.Fatalf("um %s: %v", cmd, err)
	}
	filename, err := next(l, opts.Descriptor.Val)
	if err != nil {
		log.Fatalf("um %s: %v", cmd, err)
	}
	if err := emacsNext(filename, opts.Tags.Val); err != nil {
		log.Fatalf("um %s: %v", cmd, err)
	}
}
