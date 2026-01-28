package next

import (
	"errors"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strconv"

	"github.com/brtholomy/um/go/cmd"
	"github.com/brtholomy/um/go/flags"
	"github.com/brtholomy/um/go/last"
)

const (
	CMD     = cmd.Next
	SUMMARY = "create the next um file and open with emacsclient"
)

const FILE_REGEXP = `(?m)^([0-9]+)\.[^\.]*\.*md$`

var fileRegexp *regexp.Regexp = regexp.MustCompile(FILE_REGEXP)

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

// takes the complete last file string
// returns the number as string
func NumFromLast(l string) (string, error) {
	res := fileRegexp.FindStringSubmatch(l)
	num := ""
	if len(res) < 2 {
		return num, errors.New(last.NOT_FOUND_MSG)
	}
	num = res[1]
	return num, nil
}

// takes the complete last file string and new descriptor
// returns the complete next file string
func next(last string, desc string) (string, error) {
	num, err := NumFromLast(last)
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
	opts := initOpts()
	flags.ParseArgs(CMD, SUMMARY, args, &opts)

	l, err := last.GlobLast(last.GLOB)
	if err != nil {
		log.Fatalf("um %s: %v", CMD, err)
	}
	filename, err := next(l, opts.Descriptor.Val)
	if err != nil {
		log.Fatalf("um %s: %v", CMD, err)
	}
	if err := emacsNext(filename, opts.Tags.Val); err != nil {
		log.Fatalf("um %s: %v", CMD, err)
	}
}
