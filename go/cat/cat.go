package cat

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/brtholomy/um/go/cmd"
	"github.com/brtholomy/um/go/flags"
	"github.com/brtholomy/um/go/pipe"
)

const (
	CMD            = cmd.Cat
	SUMMARY        = "cat um files together using a filelist. removes header by default"
	HR_BLOCK       = "\n---\n\n"
	DOUBLE_NEWLINE = "\n\n"
	H1             = "# "
)

type options struct {
	Filelist   flags.Arg
	Base       flags.String
	KeepHeader flags.Bool
	KeepTitle  flags.Bool
	Help       flags.Bool
}

func initOpts() options {
	return options{
		flags.Arg{"", "filelist. accepts from stdin if not provided"},
		flags.String{"--base", "-b", "", "base directory prepended to files in filelist"},
		flags.Bool{"--keep-header", "-d", false, "preserve um headers in concatenated file. overrides --keep-title"},
		flags.Bool{"--keep-title", "-t", false, "preserve um titles in concatenated file"},
		flags.Bool{"--help", "-h", false, "show help"},
	}
}

// remove the um header:
//
// # title
// : date
// + tag
//
// optionally keep just the # title
func decapitate(s string, opts options) string {
	// if there's no header at all, forget it:
	if opts.KeepHeader.IsSet() || !strings.HasPrefix(s, H1) {
		return s
	}
	// I'd rather slice and dice than mess with regex
	head, tail, ok := strings.Cut(s, DOUBLE_NEWLINE)
	if !ok {
		// head is complete if sep wasn't found:
		return head
	}
	s = tail
	if opts.KeepTitle.IsSet() {
		// we just take the first line. but when the header consists only of the title, there is no newline:
		if title, _, ok := strings.Cut(head, pipe.Newline); ok || !strings.Contains(head, pipe.Newline) {
			s = fmt.Sprintf("%s%s%s", title, DOUBLE_NEWLINE, s)
		}
	}
	return s
}

func cat(files []string, opts options) (string, error) {
	ff := make([]string, 0, len(files))
	for _, f := range files {
		bf := filepath.Join(opts.Base.Val, f)
		dat, err := os.ReadFile(bf)
		if err != nil {
			return "", fmt.Errorf("error opening target file: %w", err)
		}
		s := decapitate(string(dat), opts)
		ff = append(ff, s)
	}
	return strings.Join(ff, HR_BLOCK), nil
}

func Cat(args []string) {
	opts := initOpts()
	if err := flags.ParseArgs(CMD, SUMMARY, args, &opts); err != nil {
		var herr flags.HelpError
		if errors.As(err, &herr) {
			fmt.Println(herr)
			return
		}
		log.Fatalf("um %s: %s", CMD, err)
	}
	files, err := pipe.FileListSplitMaybeStdin(opts.Filelist.Val)
	if err != nil {
		log.Fatalf("um %s: %s", CMD, err)
	}
	s, err := cat(files, opts)
	if err != nil {
		log.Fatalf("um %s: %s", CMD, err)
	}
	fmt.Print(s)
}
